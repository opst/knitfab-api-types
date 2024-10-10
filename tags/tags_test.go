package tags_test

import (
	"encoding/json"
	"testing"

	"github.com/opst/knitfab-api-types/internal/utils/cmp"
	"github.com/opst/knitfab-api-types/tags"
)

func TestTagParsing(t *testing.T) {

	t.Run("Valid patterns", func(t *testing.T) {
		bJsonArray := []byte(
			`[
				{"key":"k1","value":"v1"}, "k1:v1",
				{"key":"k2","value":""}, "k2:",
				{"key":"","value":"v3"}, ":v3",
				{"key":"","value":""}, ":",

				"aaa:bbb:ccc",
				"aaa :bbb:ccc",
				"aaa: bbb:ccc"
			]`,
		)

		var parsedTags []tags.Tag
		if err := json.Unmarshal(bJsonArray, &parsedTags); err != nil {
			t.Fatal(err)
		}

		expectedTags := []tags.Tag{
			{Key: "k1", Value: "v1"}, {Key: "k1", Value: "v1"},
			{Key: "k2", Value: ""}, {Key: "k2", Value: ""},
			{Key: "", Value: "v3"}, {Key: "", Value: "v3"},
			{Key: "", Value: ""}, {Key: "", Value: ""},

			{Key: "aaa", Value: "bbb:ccc"},
			{Key: "aaa", Value: "bbb:ccc"},
			{Key: "aaa", Value: "bbb:ccc"},
		}

		if !cmp.SliceEqualUnordered(expectedTags, parsedTags) {
			t.Errorf(
				"did not match:\n=== expected === \n%+v\n=== actual ===\n%+v",
				expectedTags, parsedTags,
			)
		}
	})

	for name, testcase := range map[string][]byte{
		"Field 'key' is missing":           []byte(`{"keys": "k1", "value": "v1"}`),
		"Field 'key''s value is missing":   []byte(`{"key":null,"value":"v1"}`),
		"Field 'key''s value is invalid":   []byte(`{"key":[],"value":"v1"}`),
		"Field 'value' is missing":         []byte(`{"key":"k1","val":"v1"}`),
		"Field 'value''s value is missing": []byte(`{"key":"k1","value":null}`),
		"Field 'value''s value is invalid": []byte(`{"key":"k1","value":{}}`),
		"String expression without colon":  []byte(`""`),
	} {
		t.Run("Invalid pattern: "+name, func(t *testing.T) {
			var parsedTag tags.Tag
			if err := json.Unmarshal(testcase, &parsedTag); err == nil {
				t.Error("Expected error does not occured")
			}
		})
	}
}

func TestChange_unmarshal(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		payload := `{
  "add": [ "key1:value1", "key2:value2" ],
  "remove": [ "key3:value3", "key4:value4" ],
  "remove_key": [ "key5", "key6" ]
}`
		var c tags.Change
		if err := json.Unmarshal([]byte(payload), &c); err != nil {
			t.Fatal(err)
		}

		expected := tags.Change{
			AddTags: []tags.UserTag{
				{Key: "key1", Value: "value1"},
				{Key: "key2", Value: "value2"},
			},
			RemoveTags: []tags.UserTag{
				{Key: "key3", Value: "value3"},
				{Key: "key4", Value: "value4"},
			},
			RemoveKey: []string{"key5", "key6"},
		}

		if !expected.Equal(&c) {
			t.Errorf("expected: %+v, got: %+v", expected, c)
		}
	})

	t.Run("error: system tag in add", func(t *testing.T) {
		payload := `{
  "add": [ "knit#key1:value1" ],
  "remove": [ "key3:value3", "key4:value4" ],
  "remove_key": [ "key5", "key6" ]
	}`
		var c tags.Change
		if err := json.Unmarshal([]byte(payload), &c); err == nil {
			t.Error("Expected error does not occured")
		}
	})

	t.Run("error: system tag in remove", func(t *testing.T) {
		payload := `{
  "add": [ "key1:value1" ],
  "remove": [ "knit#key3:value3", "key4:value4" ],
  "remove_key": [ "key5", "key6" ]
	}`
		var c tags.Change
		if err := json.Unmarshal([]byte(payload), &c); err == nil {
			t.Error("Expected error does not occured")
		}
	})

	t.Run("error: system tag in removeKey", func(t *testing.T) {
		payload := `{
  "add": [ "key1:value1" ],
  "remove": [ "key3:value3", "key4:value4" ],
  "remove_key": [ "knit#key5", "key6" ]
	}`
		var c tags.Change
		if err := json.Unmarshal([]byte(payload), &c); err == nil {
			t.Error("Expected error does not occured")
		}
	})
}
