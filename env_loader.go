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
// a prefix. The entire name is uppercased. If a name contains a hyphen, such as with a default field
// (ie FieldA will become field-a which then becomes FIELD_A).
// string
type EnvLoader string

// Load is called internally by gonk to load a value at each node.
func (prefix EnvLoader) Load(node reflect.Value, tag tagData) (reflect.Value, error) {
	// Always create structs
	if node.Kind() == reflect.Struct {
		return reflect.New(node.Type()).Elem(), nil
	}

	// Handle actual values
	tagParts := tag.NamedKeys()
	val, ok := os.LookupEnv(prefix.toEnv(tagParts))
	if !ok {
		return reflect.Value{}, errValueNotPresent(tag, prefix)
	}
	switch node.Kind() {
	case reflect.String:
		return reflect.ValueOf(val), nil
	case reflect.Int:
		intVal, err := strconv.Atoi(val)
		if err != nil {
			return reflect.Value{}, errInvalidValue(tag, prefix)
		}
		return reflect.ValueOf(intVal), nil
	default:
		return reflect.Value{}, errValueNotSupported(tag, prefix)
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
