package gonk

import "reflect"

// MapLoader wraps a map[string]any to allow loading data from it into a struct. For a given key or
// key path it will traverse the map to the matching node, unwrap the type and set it
type MapLoader map[string]any

// Load is called when getting values from a node. It assigns a value to the passed reflect.Value, based on the tag data given.
func (m MapLoader) Load(node reflect.Value, tag tagData) error {
	val, err := m.traverse(tag)
	var out reflect.Value
	if err != nil || val == nil {
		return errValueNotPresent(tag, m)
	}
	switch node.Kind() {
	case reflect.String, reflect.Int:
		out = reflect.ValueOf(val)
	case reflect.Struct:
		out = reflect.New(node.Type()).Elem()
	case reflect.Slice:
		var sliceLen int
		switch val := val.(type) {
		case []any:
			sliceLen = len(val)
		case []map[string]any:
			sliceLen = len(val)
		default:
			return errInvalidValue(tag, m)
		}
		out = reflect.MakeSlice(node.Type(), sliceLen, sliceLen)
	case reflect.Pointer:
		return m.Load(node.Elem(), tag)
	default:
		return errValueNotSupported(tag, m)
	}
	node.Set(out)
	return nil
}

func (m MapLoader) traverse(tag tagData) (any, error) {
	// Traverse the config file
	head := any(map[string]any(m))
	for _, component := range tag.path {
		switch component.(type) {
		case string:
			// head must be a map
			headMap, ok := head.(map[string]any)
			if !ok {
				return nil, errValueNotPresent(tag, m)
			}
			head = headMap[component.(string)]
		case int:
			// head must be an array
			headSlice, ok := head.([]any)
			if !ok {
				return nil, errValueNotPresent(tag, m)
			}
			head = headSlice[component.(int)]
		}
	}
	return head, nil
}