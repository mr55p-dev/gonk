package gonk

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

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

const contents = `
---
fieldA: hello
fieldB: 10
fieldD:
  fieldE: world
fieldF:
  - fieldG: foo
    fieldH: bar
  - fieldG: baz
`

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
	}
	inp := make(map[string]any)
	assert.NoError(yaml.Unmarshal([]byte(contents), inp))
	assert.NoError(LoadConfig(
		out,
		MapLoader(inp),
	))
	assert.Equal(
		expected, *out,
		"Image was not loaded correctly",
	)
}

func TestEnvLoader(t *testing.T) {
	assert := assert.New(t)
	out := new(RootType)
	os.Setenv("CONFIG_FIELDA", "hello")
	os.Setenv("CONFIG_FIELDD_FIELDE", "world")

	expected := RootType{
		FieldA: "hello",
		FieldD: IntermediateA{
			FieldE: "world",
		},
	}
	assert.NoError(LoadConfig(
		out,
		NewEnvLoader("CONFIG"),
	))
	assert.Equal(
		expected, *out,
		"Image was not loaded correctly",
	)
}

// func TestAnotherThing(t *testing.T) {
// 	assert := assert.New(t)
// 	out := new(RootType)
// 	expected := RootType{
// 		FieldA: "hello",
// 		FieldB: 10,
// 		FieldD: IntermediateA{
// 			FieldE: "world",
// 		},
// 		FieldF: []IntermediateB{
// 			{FieldG: "foo", FieldH: "bar"},
// 			{FieldG: "baz"},
// 		},
// 	}
// 	data, _ := os.ReadFile("./test.yaml")
// 	mapData := make(map[string]any)
//
// 	assert.NoError(yaml.Unmarshal(data, mapData), "Error loading data")
// 	loader := MapLoader(mapData)
// 	errors := applyLoader(out, loader)
// 	assert.Empty(errors)
// 	assert.Equal(
// 		expected, *out,
// 		"Image was not loaded correctly",
// 	)
// }
