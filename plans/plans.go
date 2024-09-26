package plans

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/opst/knitfab-api-types/internal/utils/cmp"
	"github.com/opst/knitfab-api-types/tags"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/api/resource"
)

type Summary struct {
	// PlanId is the id of the Plan.
	PlanId string `json:"planId"`

	// Image is the container image of the Plan.
	//
	// This is exclusive with Name.
	Image *Image `json:"image,omitempty"`

	// Name is the name of the Plan.
	//
	// This is exclusive with Image, and used only for the system-builtin Plan with no image.
	Name string `json:"name,omitempty"`
}

func (s Summary) Equal(o Summary) bool {
	return s.PlanId == o.PlanId &&
		s.Image.Equal(o.Image) &&
		s.Name == o.Name
}

type Image struct {
	Repository string
	Tag        string
}

func (i *Image) Equal(o *Image) bool {
	if (i == nil) || (o == nil) {
		return (i == nil) && (o == nil)
	}
	return i.Repository == o.Repository &&
		i.Tag == o.Tag
}

// parse string as Image Tag, and upgate itself.
//
// this spec is based on docker image tag spec[^1].
//
// [^1]: https://docs.docker.com/engine/reference/commandline/tag/#description
func (i *Image) Parse(s string) error {
	// [<repository>[:<port>]/]<name>:<tag>

	ref, err := name.NewTag(s, name.WithDefaultRegistry(""))
	if err != nil {
		return err
	}

	i.Repository = ref.Repository.Name()
	i.Tag = ref.TagStr()
	return nil
}

func (i *Image) marshal() string {
	if i.Repository == "" && i.Tag == "" {
		return ""
	}
	return fmt.Sprintf(`%s:%s`, i.Repository, i.Tag)
}

func (i Image) MarshalJSON() ([]byte, error) {
	b := bytes.NewBufferString(`"`)
	b.WriteString(i.marshal())
	b.WriteString(`"`)
	return b.Bytes(), nil
}

func (i Image) MarshalYAML() (interface{}, error) {
	n := yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: i.marshal(),
		Style: yaml.DoubleQuotedStyle,
	}
	return n, nil
}

func (i *Image) UnmarshalYAML(node *yaml.Node) error {
	expr := new(string)
	err := node.Decode(expr)
	if err != nil {
		return err
	}
	return i.Parse(*expr)
}

func (i *Image) UnmarshalJSON(b []byte) error {
	expr := new(string)
	err := json.Unmarshal(b, expr)
	if err != nil {
		return err
	}
	return i.Parse(*expr)
}

func (i *Image) String() string {
	return i.marshal()
}

// Detail is the format for the response body from Knitfab APIs below:
//
// - GET  /api/plans/ (as list)
//
// - POST /api/plans/
//
// - GET  /api/plans/{planId}
//
// - PUT  /api/plans/{planId}/active
//
// - PUT  /api/plans/{planId}/resources
type Detail struct {
	Summary

	// Inputs are the input mountpoints of the plan.
	Inputs []Mountpoint `json:"inputs"`

	// Outputs are the output mountpoints of the plan.
	Outputs []Mountpoint `json:"outputs"`

	// Log is the log point of the plan.
	//
	// If nil, the plan does not record logs.
	Log *LogPoint `json:"log,omitempty"`

	// Active shows Plan's activeness.
	//
	// It is true if the plan is active and new runs can be created.
	Active bool `json:"active"`

	// OnNode is the node affinity/torelance of the plan.
	//
	// If nil, the plan does not have node affinity/torelance.
	OnNode *OnNode `json:"on_node,omitempty"`

	// Resources is the resource limits and requiremnts of the plan.
	Resources Resources `json:"resources,omitempty"`

	// ServiceAccount is the ServiceAccount name of the plan.
	//
	// Workers of the Run based this Plan will run with this ServiceAccount.
	ServiceAccount string `json:"service_account,omitempty"`
}

func (d Detail) Equal(o Detail) bool {
	logEq := d.Log == nil && o.Log == nil ||
		(d.Log != nil && o.Log != nil && d.Log.Equal(*o.Log))
	onnodeEq := d.OnNode == nil && o.OnNode == nil ||
		(d.OnNode != nil && o.OnNode != nil && d.OnNode.Equal(*o.OnNode))

	return d.Summary.Equal(o.Summary) &&
		d.Active == o.Active &&
		d.ServiceAccount == o.ServiceAccount &&
		logEq && onnodeEq &&
		cmp.MapEqual(d.Resources, o.Resources) &&
		cmp.SliceEqualUnordered(d.Inputs, o.Inputs) &&
		cmp.SliceEqualUnordered(d.Outputs, o.Outputs)
}

type Mountpoint struct {
	Path string     `json:"path"`
	Tags []tags.Tag `json:"tags"`
}

func (m Mountpoint) Equal(o Mountpoint) bool {
	return m.Path == o.Path && cmp.SliceEqualUnordered(m.Tags, o.Tags)
}

type LogPoint struct {
	Tags []tags.Tag
}

func (lp LogPoint) Equal(o LogPoint) bool {
	return cmp.SliceEqualUnordered(lp.Tags, o.Tags)
}

func (lp LogPoint) String() string {
	return fmt.Sprintf("{Tags: %+v}", lp.Tags)
}

type OnNode struct {
	May    []OnSpecLabel `json:"may,omitempty" yaml:"may,omitempty"`
	Prefer []OnSpecLabel `json:"prefer,omitempty" yaml:"prefer,omitempty"`
	Must   []OnSpecLabel `json:"must,omitempty" yaml:"must,omitempty"`
}

func (o OnNode) Equal(oo OnNode) bool {
	return cmp.SliceEqualUnordered(o.May, oo.May) &&
		cmp.SliceEqualUnordered(o.Prefer, oo.Prefer) &&
		cmp.SliceEqualUnordered(o.Must, oo.Must)
}

type OnSpecLabel struct {
	Key   string
	Value string
}

func (l OnSpecLabel) String() string {
	return fmt.Sprintf("%s=%s", l.Key, l.Value)
}

func (l OnSpecLabel) Equal(o OnSpecLabel) bool {
	return l.Key == o.Key && l.Value == o.Value
}

func (l *OnSpecLabel) Parse(s string) error {

	k, v, ok := strings.Cut(s, "=")
	if !ok {
		return fmt.Errorf("label format error (should be key=value): %s", s)
	}

	l.Key = k
	l.Value = v
	return nil
}

func (l OnSpecLabel) MarshalJSON() ([]byte, error) {
	b := bytes.NewBufferString(`"`)
	b.WriteString(l.String())
	b.WriteString(`"`)
	return b.Bytes(), nil
}

func (l OnSpecLabel) MarshalYAML() (interface{}, error) {
	n := yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: l.String(),
		Style: yaml.DoubleQuotedStyle,
	}
	return n, nil
}

func (l *OnSpecLabel) UnmarshalJSON(value []byte) error {
	expr := new(string)
	err := json.Unmarshal(value, expr)
	if err != nil {
		return err
	}
	return l.Parse(*expr)
}

func (l *OnSpecLabel) UnmarshalYAML(node *yaml.Node) error {
	expr := new(string)
	err := node.Decode(expr)
	if err != nil {
		return err
	}
	return l.Parse(*expr)
}

type Resources map[string]resource.Quantity

func (r Resources) Equal(o Resources) bool {
	return cmp.MapEqual(r, o)
}

func (r Resources) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]resource.Quantity(r))
}

func (r Resources) MarshalYAML() (interface{}, error) {
	jsonMap := map[string]string{}
	jsonBytes, err := r.MarshalJSON()
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(jsonBytes, &jsonMap)
	if err != nil {
		return nil, err
	}
	return jsonMap, nil
}

func (r *Resources) UnmarshalYAML(node *yaml.Node) error {
	var m map[string]string
	if err := node.Decode(&m); err != nil {
		return err
	}

	jsonBytes, err := json.Marshal(m)
	if err != nil {
		return err
	}
	if err := r.UnmarshalJSON(jsonBytes); err != nil {
		return err
	}

	return nil
}

func (r *Resources) UnmarshalJSON(b []byte) error {
	var m map[string]resource.Quantity
	err := json.Unmarshal(b, &m)
	if err != nil {
		return err
	}
	*r = Resources(m)
	return nil
}

// PlanSpec is the format for request body to Knitfab APIs below:
//
// - POST /api/plans/
type PlanSpec struct {
	Image          Image        `json:"image" yaml:"image"`
	Inputs         []Mountpoint `json:"inputs" yaml:"inputs"`
	Outputs        []Mountpoint `json:"outputs" yaml:"outputs"`
	Log            *LogPoint    `json:"log,omitempty" yaml:"log,omitempty"`
	OnNode         *OnNode      `json:"on_node,omitempty" yaml:"on_node,omitempty"`
	Resources      Resources    `json:"resources,omitempty" yaml:"resources,omitempty"`
	ServiceAccount string       `json:"service_account,omitempty" yaml:"service_account,omitempty"`
	Active         *bool        `json:"active" yaml:"active,omitempty"`
}

func (ps PlanSpec) Equal(o PlanSpec) bool {
	activeEq := ps.Active == nil && o.Active == nil || (ps.Active != nil && o.Active != nil && *ps.Active == *o.Active)
	logEq := ps.Log == nil && o.Log == nil || (ps.Log != nil && o.Log != nil && ps.Log.Equal(*o.Log))
	onNodeEq := ps.OnNode == nil && o.OnNode == nil || (ps.OnNode != nil && o.OnNode != nil && ps.OnNode.Equal(*o.OnNode))

	return ps.Image.Equal(&o.Image) &&
		logEq && onNodeEq && activeEq &&
		ps.ServiceAccount == o.ServiceAccount &&
		cmp.MapEqual(ps.Resources, o.Resources) &&
		cmp.SliceEqualUnordered(ps.Inputs, o.Inputs) &&
		cmp.SliceEqualUnordered(ps.Outputs, o.Outputs)
}

// ResourceLimitChange is a change of resource limit of plan.
type ResourceLimitChange struct {

	// Resource to be set.
	Set Resources `json:"set,omitempty" yaml:"set,omitempty"`

	// Resource types to be unset.
	//
	// If same type Set and Unset, Unset is affected.
	Unset []string `json:"unset,omitempty" yaml:"unset,omitempty"`
}

// SetServiceccount declares new ServiceAccount name of a Plan.
type SetServiceAccount struct {
	ServiceAccount string `json:"service_account" yaml:"service_account"`
}
