package plans

import (
	"bytes"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/opst/knitfab-api-types/v2/internal/utils/cmp"
	"github.com/opst/knitfab-api-types/v2/tags"
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

	// Entrypoint is the entrypoint of the container of the Plan.
	Entrypoint []string `json:"entrypoint,omitempty"`

	// Args are the arguments of the container of the Plan.
	Args []string `json:"args,omitempty"`

	// Name is the name of the Plan.
	//
	// This is exclusive with Image, and used only for the system-builtin Plan with no image.
	Name string `json:"name,omitempty"`

	// Annotations are the annotations of the Plan.
	//
	// In JSON format, it is a list of strings in the form of "key=value".
	Annotations Annotations `json:"annotations,omitempty"`
}

func (s Summary) Equal(o Summary) bool {
	return s.PlanId == o.PlanId &&
		s.Image.Equal(o.Image) &&
		cmp.SliceEqEq(s.Entrypoint, o.Entrypoint) &&
		cmp.SliceEqEq(s.Args, o.Args) &&
		s.Name == o.Name &&
		s.Annotations.Equal(o.Annotations)
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

type Annotations []Annotation

func (ans Annotations) Equal(o Annotations) bool {
	return cmp.SliceEqualUnordered(ans, o)
}

func (ans Annotations) marshal() []Annotation {
	_ans := append([]Annotation{}, ans...)
	slices.SortFunc(_ans, func(i, j Annotation) int {
		if c := strings.Compare(i.Key, j.Key); c != 0 {
			return c
		}
		return strings.Compare(i.Value, j.Value)
	})
	return _ans
}

func (ans Annotations) MarshalJSON() ([]byte, error) {
	s := ans.marshal()
	return json.Marshal(s)
}

type Annotation struct {
	Key   string
	Value string
}

func (an Annotation) String() string {
	return fmt.Sprintf("%s=%s", an.Key, an.Value)
}

func (an Annotation) Equal(o Annotation) bool {
	return an.Key == o.Key && an.Value == o.Value
}

func (an Annotation) MarshalJSON() ([]byte, error) {
	s := an.String()
	return json.Marshal(s)
}

func (an Annotation) MarshalYAML() (interface{}, error) {
	n := yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: an.String(),
		Style: yaml.DoubleQuotedStyle,
	}
	return n, nil
}

func (an *Annotation) parse(s string) error {
	k, v, ok := strings.Cut(s, "=")
	if !ok {
		return fmt.Errorf("annotation format error (should be key=value): %s", s)
	}

	an.Key = strings.TrimSpace(k)
	an.Value = strings.TrimSpace(v)
	return nil
}

func (an *Annotation) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	return an.parse(s)
}

func (an *Annotation) UnmarshalYAML(node *yaml.Node) error {
	var s string
	if err := node.Decode(&s); err != nil {
		return err
	}

	return an.parse(s)
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
	Inputs []Input `json:"inputs"`

	// Outputs are the output mountpoints of the plan.
	Outputs []Output `json:"outputs"`

	// Log is the log point of the plan.
	//
	// If nil, the plan does not record logs.
	Log *Log `json:"log,omitempty"`

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

// Mountpoint is the format for input/output mountpoints of a Plan.
type Mountpoint struct {
	// Path is the path of the mountpoint.
	//
	// This is the path in the container where the Data will be mounted
	// when the Run starts.
	Path string `json:"path"`

	// Tags are the tags of the mountpoint.
	//
	// For input mountpoints, these are the required tags of the Data to be mounted.
	// The Data with these all tags will be mounted to the Path when the Run starts.
	//
	// For output mountpoints, these are the tags to be attached to the Data mounted.
	Tags []tags.Tag `json:"tags"`
}

func (m Mountpoint) Equal(o Mountpoint) bool {
	return m.Path == o.Path && cmp.SliceEqualUnordered(m.Tags, o.Tags)
}

// Upstream is the format for input dependencies of a Plan.
type Upstream struct {
	// Plan is the upstream Plan.
	Plan Summary `json:"plan"`

	// Mountpoint represents the Output which is directt upstream.
	//
	// Log and Mountpoint are mutually exclusive.
	Mountpoint *Mountpoint `json:"mountpoint,omitempty"`

	// Log represents the Log which is direct upstream.
	//
	// Log and Mountpoint are mutually exclusive.
	Log *LogPoint `json:"log,omitempty"`
}

func (d Upstream) Equal(o Upstream) bool {
	if (d.Mountpoint == nil) != (o.Mountpoint == nil) {
		return false
	}

	if (d.Log == nil) != (o.Log == nil) {
		return false
	}

	mountpointMatch := (d.Mountpoint == nil && o.Mountpoint == nil) ||
		d.Mountpoint.Equal(*o.Mountpoint)

	logMatch := (d.Log == nil && o.Log == nil) ||
		d.Log.Equal(*o.Log)

	return d.Plan.Equal(o.Plan) && mountpointMatch && logMatch
}

// Input is the format for input mountpoints of a Plan.
type Input struct {
	Mountpoint

	// Upstreams are the upstream Plans and their output mountpoints
	// whose output Data can be mounted to this input mountpoint.
	Upstreams []Upstream `json:"upstreams"`
}

func (i Input) Equal(o Input) bool {
	return i.Mountpoint.Equal(o.Mountpoint) &&
		cmp.SliceEqualUnordered(i.Upstreams, o.Upstreams)
}

// Downstream is the format for output dependencies of a Plan.
type Downstream struct {
	// Plan is the downstream Plan.
	Plan Summary `json:"plan"`

	// Mountpoint represents the Input which is direct downstream.
	Mountpoint Mountpoint `json:"mountpoint"`
}

func (d Downstream) Equal(o Downstream) bool {
	return d.Plan.Equal(o.Plan) && d.Mountpoint.Equal(o.Mountpoint)
}

// Output is the format for output mountpoints of a Plan.
type Output struct {
	Mountpoint

	// Downstreams are the downstream Plans and their input mountpoints
	// can be assigned with Data from this output.
	Downstreams []Downstream `json:"downstreams"`
}

func (o Output) Equal(oo Output) bool {
	return o.Mountpoint.Equal(oo.Mountpoint) &&
		cmp.SliceEqualUnordered(o.Downstreams, oo.Downstreams)
}

type Log struct {
	LogPoint

	// Downstreams are the downstream Plans and their input mountpoints
	// can be assigned with Data from this output.
	Downstreams []Downstream `json:"downstreams"`
}

func (l Log) Equal(ol Log) bool {
	return l.LogPoint.Equal(ol.LogPoint) &&
		cmp.SliceEqualUnordered(l.Downstreams, ol.Downstreams)
}

func (l Log) String() string {
	return fmt.Sprintf("{LogPoint: %+v, Downstreams: %+v}", l.LogPoint, l.Downstreams)
}

type LogPoint struct {
	Tags []tags.Tag `json:"tags"`
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
	// Annotations are the annotations of the Plan.
	//
	// In JSON format, it is a list of strings in the form of "key=value".
	//
	// If same key is set multiple times, the last one is used.
	Annotations Annotations `json:"annotations,omitempty" yaml:"annotations,omitempty"`

	// Image is the container image of the Plan.
	Image Image `json:"image" yaml:"image"`

	// Entrypoint is the entrypoint of the container of the Plan.
	Entrypoint []string `json:"entrypoint,omitempty" yaml:"entrypoint,omitempty"`

	// Args are the arguments of the container of the Plan.
	Args []string `json:"args,omitempty" yaml:"args,omitempty"`

	// Inputs are the input mountpoints of the plan.
	//
	// These describes "where should input Data be mounted to the container" and
	// "what tags should be attached to the input Data".
	//
	// When Knitfab detect the Data with the tags, it will be mounted to the container and started as a Run.
	Inputs []Mountpoint `json:"inputs" yaml:"inputs"`

	// Outputs are the output mountpoints of the plan.
	//
	// These describes "where should output Data be mounted to the container" and
	// "what tags will be attached to the output Data".
	//
	// On the Run start, Knitfab will create the mountpoints and attach the tags to the output Data.
	Outputs []Mountpoint `json:"outputs" yaml:"outputs"`

	// Log is the log point of the plan.
	//
	// "Log" means a Data containing standard output and standard error of the container.
	// If nil, the plan does not record logs.
	//
	// As Outputs, Log can have Tags which will be attached to the log Data.
	Log *LogPoint `json:"log,omitempty" yaml:"log,omitempty"`

	// OnNode is the node affinity/torelance of the plan.
	//
	// If nil, the plan does not have node affinity/torelance.
	OnNode *OnNode `json:"on_node,omitempty" yaml:"on_node,omitempty"`

	// Resources is the conputational resource limits and requiremnts of the plan.
	Resources Resources `json:"resources,omitempty" yaml:"resources,omitempty"`

	// ServiceAccount is the Kubernetes ServiceAccount name of the plan.
	ServiceAccount string `json:"service_account,omitempty" yaml:"service_account,omitempty"`

	// Active shows Plan's activeness.
	//
	// If true or nil, the Plan is active and new Runs based the Plan can be started.
	//
	// If false, the Plan is inactive and new Runs based the Plan are created but suspended to start.
	Active *bool `json:"active" yaml:"active,omitempty"`
}

func (ps PlanSpec) Equal(o PlanSpec) bool {
	logEq := ps.Log == nil && o.Log == nil || (ps.Log != nil && o.Log != nil && ps.Log.Equal(*o.Log))
	onNodeEq := ps.OnNode == nil && o.OnNode == nil || (ps.OnNode != nil && o.OnNode != nil && ps.OnNode.Equal(*o.OnNode))
	activeEq := ps.Active == nil && o.Active == nil || (ps.Active != nil && o.Active != nil && *ps.Active == *o.Active)

	return ps.Annotations.Equal(o.Annotations) &&
		ps.Image.Equal(&o.Image) &&
		cmp.SliceEqEq(ps.Entrypoint, o.Entrypoint) &&
		cmp.SliceEqEq(ps.Args, o.Args) &&
		cmp.SliceEqualUnordered(ps.Inputs, o.Inputs) &&
		cmp.SliceEqualUnordered(ps.Outputs, o.Outputs) &&
		logEq &&
		onNodeEq &&
		cmp.MapEqual(ps.Resources, o.Resources) &&
		ps.ServiceAccount == o.ServiceAccount &&
		activeEq
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

// AnnotationChange is a changeset of Annotations of a Plan.
//
// Knitfab WebAPI applies Remove first, then Add.
type AnnotationChange struct {
	// Annotations to be added.
	//
	// If the Plan to be annotated already has the key, the value is updated.
	// If same key is set multiple times, the last one is used.
	Add Annotations `json:"add,omitempty" yaml:"add,omitempty"`

	// Keys of Annotations to be removed.
	Remove Annotations `json:"remove,omitempty" yaml:"remove,omitempty"`

	RemoveKey []string `json:"remove_key,omitempty" yaml:"remove_key,omitempty"`
}
