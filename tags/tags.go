package tags

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/opst/knitfab-api-types/v2/internal/utils/cmp"
	"github.com/opst/knitfab-api-types/v2/misc/rfctime"
	"gopkg.in/yaml.v3"
)

const (
	SystemTagPrefix  string = "knit#"
	KeyKnitId        string = SystemTagPrefix + "id"
	KeyKnitTimestamp string = SystemTagPrefix + "timestamp"
	KeyKnitTransient string = SystemTagPrefix + "transient"

	// ValueKnitTransientFailed is the value of KeyKnitTransient
	// when the upstream Run of the Data is failed.
	ValueKnitTransientFailed string = "failed"

	// ValueKnitTransientProcessing is the value of KeyKnitTransient
	// when the upstream Run of the Data is processing and the Data is under creation.
	ValueKnitTransientProcessing string = "processing"

	// ValueKnitTransientPurged is the value of KeyKnitTransient
	// when the Data is purged.
	// Any Runs using the Data cannot be retried, and the Data cannot be downloaded.
	ValueKnitTransientPurged string = "purged"
)

// Tag represents Tag for Data and Plan input/output.
//
// To make this type from user inputted value, use Tag.Parse method.
type Tag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (t Tag) String() string {
	return t.Key + ":" + t.Value
}

func (a Tag) Equal(b Tag) bool {
	if a.Key != b.Key {
		return false
	}

	if a.Key != KeyKnitTimestamp {
		return a.Value == b.Value
	}

	vA, errA := rfctime.ParseRFC3339DateTime(a.Value)
	vB, errB := rfctime.ParseRFC3339DateTime(b.Value)

	return (errA == nil) && (errB == nil) &&
		vA.Equiv(vB)
}

// parse and validation string value as Tag
//
// # Args
//
// - string: "KEY:VALUE" formatted string. If not, it returns error.
func (t *Tag) Parse(s string) error {
	k, v, ok := strings.Cut(s, ":")
	if !ok {
		return fmt.Errorf("tag parse error: %s :no key", s)
	}

	k = strings.TrimSpace(k)
	v = strings.TrimSpace(v)

	switch k {
	case KeyKnitTimestamp:
		_, err := rfctime.ParseRFC3339DateTime(v)
		if err != nil {
			return fmt.Errorf("tag parse error: %s is not timestamp", s)
		}
	case KeyKnitTransient:
		switch v {
		case ValueKnitTransientProcessing, ValueKnitTransientFailed, ValueKnitTransientPurged:
			// pass
		default:
			return fmt.Errorf(
				`tag parse error: "%s" should be one of "%s", "%s", or "%s"`,
				KeyKnitTransient, ValueKnitTransientProcessing, ValueKnitTransientFailed, ValueKnitTransientPurged,
			)
		}
	}
	t.Key = k
	t.Value = v

	return nil
}

// UserTag represents user specified Tag for Data and Plan input/output.
//
// To make this type from user inputted value, use UserTag.Parse method or Tag.AsUserTag method.
type UserTag Tag

// AsUserTag returns true if the tag is not system tag.
//
// # Args
//
// - ut: UserTag to be filled.
//
// # Returns
//
// - bool: true if the tag is not system tag.
func (t Tag) AsUserTag(ut *UserTag) bool {
	if strings.HasPrefix(t.Key, SystemTagPrefix) {
		return false
	}
	*ut = UserTag(t)
	return true
}

func (t *Tag) UnmarshalJSON(data []byte) error {
	{
		s := new(string)
		if err := json.Unmarshal(data, s); err == nil {
			return t.Parse(*s)
		}
	}

	var dat map[string]interface{}
	if err := json.Unmarshal(data, &dat); err != nil {
		return errors.New(`failed to parse Tag`)
	}

	return t.unarshal(dat)
}

func (t *Tag) UnmarshalYAML(n *yaml.Node) error {
	{
		s := new(string)
		if err := n.Decode(s); err == nil {
			return t.Parse(*s)
		}
	}

	var dat map[string]interface{}
	if err := n.Decode(&dat); err != nil {
		return errors.New(`failed to parse Tag`)
	}
	return t.unarshal(dat)
}

func (t Tag) marshal() string {
	return t.String()
}

func (ut Tag) MarshalJSON() ([]byte, error) {
	return []byte(`"` + ut.marshal() + `"`), nil
}

func (ut Tag) MarshalYAML() (interface{}, error) {
	n := yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: ut.marshal(),
		Style: yaml.DoubleQuotedStyle,
	}
	return n, nil
}

// parse and validation string value as UserTag
//
// # Args
//
// - string: "KEY:VALUE" formatted string. If not, it returns error.
// If KEY part is started with "knit#", it returns error.
func (ut *UserTag) Parse(s string) error {
	t := &Tag{}
	if err := t.Parse(s); err != nil {
		return err
	}
	if strings.HasPrefix(t.Key, SystemTagPrefix) {
		return fmt.Errorf(`tag key "%s..." is reserved for system tags`, SystemTagPrefix)
	}
	*ut = UserTag(*t)
	return nil
}

func (t *Tag) unarshal(dat map[string]interface{}) error {
	if dat == nil {
		return errors.New("tag is nil")
	}

	// check key
	bkey, ok := dat["key"]
	if !ok {
		return errors.New(`field "key" is missing`)
	}
	if bkey == nil {
		return errors.New(`field "key"'s value is missing`)
	}
	key, ok := bkey.(string)
	if !ok {
		return errors.New(`field "key"'s value is invalid`)
	}
	t.Key = key

	// check value
	bvalue, ok := dat["value"]
	if !ok {
		return errors.New(`field "value" is missing`)
	}
	if bvalue == nil {
		return errors.New(`field "value"'s value is missing`)
	}
	value, ok := bvalue.(string)
	if !ok {
		return errors.New(`field "value"'s value is invalid`)
	}
	t.Value = value

	return nil
}

func (ut *UserTag) UnmarshalJSON(data []byte) error {
	t := &Tag{}
	if err := t.UnmarshalJSON(data); err != nil {
		return err
	}
	if strings.HasPrefix(t.Key, SystemTagPrefix) {
		return fmt.Errorf(`tag key "%s..." is reserved for system tags`, SystemTagPrefix)
	}
	*ut = UserTag(*t)
	return nil
}

func (u UserTag) Equal(o UserTag) bool {
	ut, ot := Tag(u), Tag(o)
	return ut.Equal(ot)
}

// Change is the format for request body to change tags.
//
// This type is used for:
//
// - POST /api/data/{knitId}
type Change struct {
	AddTags    []UserTag `json:"add"`
	RemoveTags []UserTag `json:"remove"`
	RemoveKey  []string  `json:"remove_key"`
}

func (c *Change) UnmarshalJSON(data []byte) error {

	type raw struct {
		AddTags    []UserTag `json:"add"`
		RemoveTags []UserTag `json:"remove"`
		RemoveKey  []string  `json:"remove_key"`
	}

	var r raw
	if err := json.Unmarshal(data, &r); err != nil {
		return err
	}

	for _, rk := range r.RemoveKey {
		if strings.HasPrefix(rk, SystemTagPrefix) {
			return fmt.Errorf(`tag key "%s..." is reserved for system tags. not removable.`, SystemTagPrefix)
		}
	}

	c.AddTags = r.AddTags
	c.RemoveTags = r.RemoveTags
	c.RemoveKey = r.RemoveKey
	return nil
}

func (c *Change) Equal(o *Change) bool {

	return cmp.SliceEqualUnordered(c.AddTags, o.AddTags) &&
		cmp.SliceEqualUnordered(c.RemoveTags, o.RemoveTags) &&
		cmp.SliceEqEqUnordered(c.RemoveKey, o.RemoveKey)
}
