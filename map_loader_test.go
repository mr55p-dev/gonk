package gonk

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTraverseMap(t *testing.T) {
	assert := assert.New(t)
	testMap := MapLoader(map[string]any{
		"key": "value",
		"nested_key_1": map[string]any{
			"nested_key_2": map[string]any{
				"nested_value": "nested_value",
				"nested_int":   2,
			},
		},
	})
	val, err := testMap.traverse(
		tagData{
			path: []any{
				"nested_key_1",
				"nested_key_2",
				"nested_value",
			},
		},
	)
	assert.NoError(err)
	assert.Equal("nested_value", val, "key should be nested_value")

	val, err = testMap.traverse(
		tagData{
			path: []any{
				"nested_key_1",
				"nested_key_2",
				"nested_int",
			},
		},
	)
	assert.Nil(err)
	assert.Equal(2, val)
}

func TestNoKeyTraverseMap(t *testing.T) {
	assert := assert.New(t)
	testMap := MapLoader(map[string]any{
		"key": "value",
		"nested_key_1": map[string]any{
			"nested_key_2": map[string]any{
				"nested_int": 2,
			},
		},
	})
	val, err := testMap.traverse(
		tagData{
			path: []any{
				"nested_value",
				"nested_key_1",
				"nested_key_2",
			},
		},
	)
	assert.ErrorContains(err, "nested_value")
	assert.IsType(ValueNotPresentError(""), err)
	assert.Zero(val)
}

func TestMapLoader(t *testing.T) {
	assert := assert.New(t)
	out := new(RootType)
	expected := RootType{
		FieldA: "hello",
		FieldB: 10,
		FieldD: IntermediateA{
			FieldE: "world",
		},
		FieldF: []IntermediateB{
			{FieldG: "foo", FieldH: "bar"},
			{FieldG: "baz"},
		},
		FieldH: "-",
		FieldI: true,
	}

	inp := MapLoader(map[string]any{
		"fieldA": "hello",
		"fieldB": 10,
		"fieldD": map[string]any{
			"fieldE": "world",
		},
		"fieldF": []any{
			map[string]any{"fieldG": "foo", "fieldH": "bar"},
			map[string]any{"fieldG": "baz"},
		},
		"field-h": "-",
		"field-i": true,
	})
	assert.NoError(LoadConfig(out, inp))
	assert.Equal(expected, *out)
}
