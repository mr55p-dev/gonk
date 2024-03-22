package gonk

import (
	"os"
	"reflect"
	"strconv"
	"strings"
)

// EnvLoader will load configuration from environment variables. Supports only string and int types.
// Does not support loading slice elements. Nested data will concatenate all previous path elements
// into one env name, separated by '_'. If the undeerlying string is not "", this will be used as
// a prefix. The entire name is uppercased.
// string
type EnvLoader string

// Load is called internally by gonk to load a value at each node.
func (prefix EnvLoader) Load(node reflect.Value, tag tagData) error {
	// Always create structs
	if node.Kind() == reflect.Struct {
		val := reflect.New(node.Type()).Elem()
		node.Set(val)
		return nil
	}

	// Handle actual values
	tagParts := tag.NamedKeys()
	val, ok := os.LookupEnv(prefix.toEnv(tagParts))
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

func (prefix EnvLoader) toEnv(parts []string) string {
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
