package gonk

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
	}

	os.Setenv("CONFIG_FIELDB", "10")
	os.Setenv("CONFIG_FIELDD_FIELDE", "world")

	mapLoader := MapLoader(map[string]any{
		"fieldA": "hello",
		"fieldD": map[string]any{},
		"fieldF": []any{
			map[string]any{"fieldG": "foo", "fieldH": "bar"},
			map[string]any{"fieldG": "baz"},
		},
	})
	envLoader := EnvLoader("config")
	assert.NoError(LoadConfig(out, mapLoader, envLoader))
	assert.Equal(expected, *out)
}
