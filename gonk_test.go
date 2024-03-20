package gonk

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTraverseMap(t *testing.T) {
	assert := assert.New(t)
	testMap := map[string]any{
		"key": "value",
		"nested_key_1": map[string]any{
			"nested_key_2": map[string]any{
				"nested_value": "nested_value",
				"nested_int":   2,
			},
		},
	}
	val, err := traverseMap(
		testMap,
		"nested_value",
		"nested_key_1",
		"nested_key_2",
	)
	assert.Nil(err, "there should be no error")
	assert.Equal("nested_value", val, "key should be nested_value")

	val, err = traverseMap(
		testMap,
		"nested_int",
		"nested_key_1",
		"nested_key_2",
	)
	assert.Nil(err)
	assert.Equal(val, 2)
}

func TestNoKeyTraverseMap(t *testing.T) {
	assert := assert.New(t)
	testMap := map[string]any{
		"key": "value",
		"nested_key_1": map[string]any{
			"nested_key_2": map[string]any{
				"nested_int": 2,
			},
		},
	}
	val, err := traverseMap(
		testMap,
		"nested_value",
		"nested_key_1",
		"nested_key_2",
	)
	assert.ErrorContains(err, "nested_value")
	assert.IsType(&KeyNotPresent{}, err)
	assert.Zero(val)
}

func TestGetEnvName(t *testing.T) {
	assert := assert.New(t)
	tests := map[string]string{
		"key":             "KEY",
		"key-with-hyphen": "KEY_WITH_HYPHEN",
		"key_with-mixed":  "KEY_WITH_MIXED",
		"path.key-name":   "PATH_KEY_NAME",
	}
	for k, v := range tests {
		assert.Equal(v, getEnvName(k, ""))
	}
}

func TestGetEnvNameWithPrefix(t *testing.T) {
	assert := assert.New(t)
	tests := map[string]string{
		"key":           "XYZ_KEY",
		"path.key-name": "XYZ_PATH_KEY_NAME",
	}
	for k, v := range tests {
		assert.Equal(v, getEnvName(k, "xyz"))
	}
}

func TestParseTag(t *testing.T) {
	assert := assert.New(t)
	config := "path.segment.key"
	tag := parseConfigTag(config)
	assert.Equal("key", tag.key)
	assert.Equal([]string{"path", "segment"}, tag.path)
	assert.Equal(config, tag.config)
	assert.False(tag.options.optional)
}

func TestParseTagOptional(t *testing.T) {
	assert := assert.New(t)
	config := "path.segment.key,optional"
	tag := parseConfigTag(config)
	assert.Equal("key", tag.key)
	assert.Equal([]string{"path", "segment"}, tag.path)
	assert.Equal(config, tag.config)
	assert.True(tag.options.optional)
}

type IntermediateA struct {
	FieldE string `config:"fieldE"`
}

type IntermediateB struct {
	FieldG string `config:"fieldG"`
	FieldH string `config:"fieldH,optional"`
}

type RootType struct {
	FieldA string          `config:"fieldA"`
	FieldB int             `config:"fieldB,optional"`
	FieldC string          `config:"fieldC,optional"`
	FieldD IntermediateA   `config:"fieldD"`
	FieldF []IntermediateB `config:"fieldF,optional"`
}

func TestSomething(t *testing.T) {
	assert := assert.New(t)
	out := new(RootType)
	expected := RootType{
		FieldA: "hello",
		FieldB: 10,
		FieldD: IntermediateA{
			FieldE: "world",
		},
		FieldF: []IntermediateB{
			{
				FieldG: "foo",
				FieldH: "bar",
			},
			{
				FieldG: "baz",
			},
		},
	}
	assert.NoError(LoadConfig(
		out,
		FileLoader("./test.yaml", false),
	))
	assert.Equal(
		expected, *out,
		"Image was not loaded correctly",
	)
}

func TestSomethingElse(t *testing.T) {
	assert := assert.New(t)
	out := new(RootType)
	os.Setenv("CONFIG_FIELDA", "hello")
	os.Setenv("CONFIG_FIELDC", "hello")
	os.Setenv("CONFIG_FIELDD_FIELDE", "world")

	expected := RootType{
		FieldA: "hello",
		FieldD: IntermediateA{
			FieldE: "world",
		},
	}
	assert.NoError(LoadConfig(
		out,
		EnvironmentLoader("CONFIG"),
	))
	assert.Equal(
		expected, *out,
		"Image was not loaded correctly",
	)
}
