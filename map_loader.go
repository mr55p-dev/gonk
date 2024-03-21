package gonk

import (
	"fmt"
	"reflect"
)

type MapLoader map[string]any

func (m MapLoader) Set(node reflect.Value, tag Tag) (reflect.Value, error) {
	zero := reflect.Zero(node.Type())
	val, err := traverse(map[string]any(m), tag)
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
		val, err := traverse(map[string]any(m), tag)
		if err != nil {
			return zero, errKeyNotPresent(tag, m)
		}
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
