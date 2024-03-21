package gonk

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

type EnvLoader string

func NewEnvLoader(envPrefix string) Loader {
	return EnvLoader(envPrefix)
}

func (prefix EnvLoader) Set(node reflect.Value, tag Tag) (reflect.Value, error) {
	// Always create structs
	if node.Kind() == reflect.Struct {
		val := reflect.New(node.Type()).Elem()
		return val, nil
	}

	// Handle actual values
	zero := reflect.Zero(node.Type())
	tagParts := tag.NamedKeys()
	val, ok := os.LookupEnv(prefix.ToEnv(tagParts))
	if !ok {
		return zero, errKeyNotPresent(tag, prefix)
	}
	switch node.Kind() {
	case reflect.String:
		return reflect.ValueOf(val), nil
	case reflect.Int:
		intVal, err := strconv.Atoi(val)
		if err != nil {
			return zero, errInvalidValue(tag, prefix)
		}
		return reflect.ValueOf(intVal), nil
	default:
		return zero, fmt.Errorf("Invalid key type for key %s", tag)
	}
}

func (prefix EnvLoader) ToEnv(parts []string) string {
	replacer := strings.NewReplacer(
		"-", "_",
		".", "_",
	)
	envParts := []string{}
	if string(prefix) != "" {
		envParts = append(envParts, string(prefix))
	}
	envParts = append(envParts, parts...)
	out := strings.Join(envParts, "_")
	out = strings.ToUpper(out)
	out = replacer.Replace(out)
	return out
}
