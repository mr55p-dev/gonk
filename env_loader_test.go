package gonk

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetEnvName(t *testing.T) {
	assert := assert.New(t)
	tests := map[string]string{
		"key":             "KEY",
		"key-with-hyphen": "KEY_WITH_HYPHEN",
		"key_with-mixed":  "KEY_WITH_MIXED",
		"path.key-name":   "PATH_KEY_NAME",
	}
	loader := EnvLoader("")
	for input, expected := range tests {
		tag := parseConfigTag(input)
		assert.Equal(expected, loader.ToEnv(tag.NamedKeys()))
	}
}

func TestGetEnvNameWithPrefix(t *testing.T) {
	assert := assert.New(t)
	tests := map[string]string{
		"key":           "XYZ_KEY",
		"path.key-name": "XYZ_PATH_KEY_NAME",
	}
	loader := EnvLoader("xyz")
	for input, expected := range tests {
		tag := parseConfigTag(input)
		assert.Equal(expected, loader.ToEnv(tag.NamedKeys()))
	}
}

func TestEnvLoader(t *testing.T) {
	assert := assert.New(t)
	out := new(RootType)
	os.Setenv("CONFIG_FIELDA", "hello")
	os.Setenv("CONFIG_FIELDB", "10")
	os.Setenv("CONFIG_FIELDD_FIELDE", "world")

	expected := RootType{
		FieldA: "hello",
		FieldB: 10,
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
