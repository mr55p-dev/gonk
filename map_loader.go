package gonk

import (
	"fmt"
	"os"
	"reflect"

	"gopkg.in/yaml.v3"
)

type MapLoader map[string]any

func NewMapLoader(data map[string]any) Loader {
	return MapLoader(data)
}

func (m MapLoader) Set(node reflect.Value, tag Tag) (reflect.Value, error) {
	zero := reflect.Zero(node.Type())
	val, err := m.traverse(tag)
	if err != nil || val == nil {
		return zero, errKeyNotPresent(tag, m)
	}
	switch node.Kind() {
	case reflect.String, reflect.Int:
		return reflect.ValueOf(val), nil
	case reflect.Struct:
		val := reflect.New(node.Type()).Elem()
		return val, nil
	case reflect.Slice:
		valSlice, ok := val.([]any)
		if !ok {
			return zero, errInvalidValue(tag, m)
		}
		slice := reflect.MakeSlice(node.Type(), len(valSlice), len(valSlice))
		return slice, nil
	case reflect.Pointer:
		return m.Set(node.Elem(), tag)
	default:
		return zero, fmt.Errorf("Invalid type for key %s", tag)
	}
}

func (m MapLoader) traverse(tag Tag) (any, error) {
	// Traverse the config file
	head := any(map[string]any(m))
	for _, component := range tag.path {
		switch component.(type) {
		case string:
			// head must be a map
			headMap, ok := head.(map[string]any)
			if !ok {
				return nil, errKeyNotPresent(tag, m)
			}
			head = headMap[component.(string)]
		case int:
			// head must be an array
			headSlice, ok := head.([]any)
			if !ok {
				return nil, errKeyNotPresent(tag, m)
			}
			head = headSlice[component.(int)]
		}
	}
	return head, nil
}

func loadYamlFile(filename string) (map[string]any, error) {
	out := make(map[string]any)
	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(file, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}
