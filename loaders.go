package gonk

import (
	"reflect"
)

type Loader interface {
	Set(node reflect.Value, tag Tag) (reflect.Value, error)
}

func Queue(node reflect.Value, tag Tag) (out []*Node, err error) {
	switch node.Kind() {
	case reflect.Struct:
		nodeType := node.Type()
		for i := 0; i < nodeType.NumField(); i++ {
			newFrame := new(Node)
			fieldType := nodeType.Field(i)
			tagRaw, ok := fieldType.Tag.Lookup("config")
			if !ok {
				continue
			}

			newFrame.valueOf = node.Field(i)
			newFrame.tag = parseConfigTag(tagRaw)
			newFrame.tag = tag.Push(newFrame.tag)
			out = append(out, newFrame)
		}
		return
	case reflect.Slice:
		for i := 0; i < node.Len(); i++ {
			frame := new(Node)
			frame.valueOf = node.Index(i)
			frame.tag = tag.Push(Tag{
				path: []any{i},
			})
			out = append(out, frame)
		}
		return
	case reflect.Pointer:
		return Queue(node.Elem(), tag)
	default:
		return
	}
}

func applyLoader(target any, l Loader) errorList {
	errs := make(errorList, 0)
	nodeStk := new(Stack)
	frames, err := Queue(reflect.ValueOf(target), Tag{})
	if err != nil {
		errs = append(errs, err)
		return errs
	}
	for _, v := range frames {
		nodeStk.Push(v)
	}
	for nodeStk.Size() > 0 {
		node := nodeStk.Pop()
		// Set the nodes value
		newVal, err := l.Set(node.valueOf, node.tag)
		if err != nil {
			switch err.(type) {
			case KeyNotPresent:
				if !node.tag.options.optional {
					errs = append(errs, err)
				}
			default:
				errs = append(errs, err)
			}
		}
		node.valueOf.Set(newVal)

		// Queue new nodes from it if needed
		frames, err := Queue(node.valueOf, node.tag)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		for _, frame := range frames {
			nodeStk.Push(frame)
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return errs
}

func FileLoader(configFile string) Loader {
	// file, err := loadYamlFile(configFile)
	return NewMapLoader(make(map[string]any))
}
