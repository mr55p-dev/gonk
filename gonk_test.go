package gonk

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
	FieldG string          `config:"-"`
	FieldH string
}

func TestMultiLoader(t *testing.T) {
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
		FieldH: "bin",
	}

	t.Setenv("CONFIG_FIELDB", "10")
	t.Setenv("CONFIG_FIELDD_FIELDE", "world")

	mapLoader := MapLoader(map[string]any{
		"fieldA": "hello",
		"fieldD": map[string]any{},
		"fieldF": []any{
			map[string]any{"fieldG": "foo", "fieldH": "bar"},
			map[string]any{"fieldG": "baz"},
		},
		"field-g": "don't look at me",
		"field-h": "bin",
	})
	envLoader := EnvLoader("config")
	assert.NoError(LoadConfig(out, mapLoader, envLoader, nil))
	assert.Equal(expected, *out)
}
