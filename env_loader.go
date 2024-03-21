package gonk

import (
	"os"
	"reflect"
	"strconv"
	"strings"
)

type EnvLoader string

func NewEnvLoader(envPrefix string) Loader {
	return EnvLoader(envPrefix)
}

func (prefix EnvLoader) GetValue(node reflect.Value, tag Tag) error {
	// Always create structs
	if node.Kind() == reflect.Struct {
		val := reflect.New(node.Type()).Elem()
		node.Set(val)
		return nil
	}

	// Handle actual values
	tagParts := tag.NamedKeys()
	val, ok := os.LookupEnv(prefix.ToEnv(tagParts))
	if !ok {
		return errValueNotPresent(tag, prefix)
	}
	switch node.Kind() {
	case reflect.String:
		node.Set(reflect.ValueOf(val))
		return nil
	case reflect.Int:
		intVal, err := strconv.Atoi(val)
		if err != nil {
			return errInvalidValue(tag, prefix)
		}
		node.Set(reflect.ValueOf(intVal))
		return nil
	default:
		return errValueNotSupported(tag, prefix)
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
