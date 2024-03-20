package gonk

import (
	"errors"
	"fmt"
	"os"
	"reflect"
)

type Loader func(dest any) errorList

func nilLoaderFn(dest any) errorList {
	return nil
}

type StackFrame struct {
	tag     Tag
	valueOf reflect.Value

	typeOf reflect.Type
	data   any
}

type Stack struct {
	storage []*StackFrame
}

func (s *Stack) Push(frame *StackFrame) {
	s.storage = append(s.storage, frame)
}

func (s *Stack) Pop() *StackFrame {
	if len(s.storage) == 0 {
		return nil
	}
	end := len(s.storage) - 1
	out := s.storage[end]
	s.storage = s.storage[:end]
	return out
}

func (s *Stack) Size() int {
	return len(s.storage)
}

// func queueStruct(stack *Stack, data map[string]any, baseFrame *StackFrame) {
// 	for i := 0; i < baseFrame.typeOf.NumField(); i++ {
// 		// Create a new stack frame
// 		frame := new(StackFrame)
// 		frame.typeOf = baseFrame.typeOf.Field(i).Type
// 		frame.valueOf = baseFrame.valueOf.Field(i)
// 		// Parse the field tag
// 		tagRaw, ok := baseFrame.typeOf.Field(i).Tag.Lookup("config")
// 		if !ok {
// 			continue
// 		}
// 		frame.tag = parseConfigTag(tagRaw)
// 		if baseFrame.tag.kavafey != "" {
// 			frame.tag = tagPathConcat(frame.tag, baseFrame.tag.key)
// 		}
// 		frame.data = data[frame.tag.key]
// 		stack.Push(frame)
// 	}
// 	return
// }

func queueSlice(stack *Stack, arrData []any, elem *StackFrame) {
	for idx, arrElem := range arrData {
		newFrame := new(StackFrame)
		newFrame.data = arrElem
		newFrame.typeOf = elem.typeOf.Elem()
		newFrame.valueOf = elem.valueOf.Index(idx)
		stack.Push(newFrame)
	}
	return
}

type loader interface {
	Set(node reflect.Value, tag Tag) (reflect.Value, error)
	Queue(node reflect.Value, tag Tag) ([]*StackFrame, error)
}

type mapLoader map[string]any
type envLoader string

func (m mapLoader) Set(node reflect.Value, tag Tag) (reflect.Value, error) {
	zero := reflect.Zero(node.Type())
	val, err := traverse(map[string]any(m), tag)
	if err != nil || val == nil {
		return zero, errKeyNotPresent(tag.String())
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
			return zero, errKeyNotPresent(tag.String())
		}
		valSlice, ok := val.([]any)
		if !ok {
			return zero, errInvalidValue(tag.String())
		}
		slice := reflect.MakeSlice(node.Type(), len(valSlice), len(valSlice))
		return slice, nil
	case reflect.Pointer:
		return m.Set(node.Elem(), tag)
	default:
		return zero, fmt.Errorf("Invalid type for key %s", tag.String())
	}
}

func (prefix envLoader) Set(node reflect.Value, tag Tag) (reflect.Value, error) {
	zero := reflect.Zero(node.Type())
	switch node.Kind() {
	case reflect.String:
		val, ok := os.LookupEnv(prefix.getEnvName(tag))
		if !ok {
			return zero, errKeyNotPresent(tag.String())
		}
		return reflect.ValueOf(val), nil
	case reflect.Struct:
		val := reflect.New(node.Type()).Elem()
		return val, nil
	default:
		return zero, fmt.Errorf("Invalid tkey type for key %s", tag.String())
	}
}

func (m mapLoader) Queue(node reflect.Value, tag Tag) (out []*StackFrame, err error) {
	switch node.Kind() {
	case reflect.Struct:
		nodeType := node.Type()
		for i := 0; i < nodeType.NumField(); i++ {
			newFrame := new(StackFrame)
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
			frame := new(StackFrame)
			frame.valueOf = node.Index(i)
			frame.tag = tag.Push(Tag{
				path: []any{i},
			})
			out = append(out, frame)
		}
		return
	case reflect.Pointer:
		return m.Queue(node.Elem(), tag)
	default:
		return
	}
}

func GenericLoader(target any, l loader) error {
	errs := make(errorList, 0)
	nodeStk := new(Stack)
	frames, err := l.Queue(reflect.ValueOf(target), Tag{})
	if err != nil {
		return err
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
			case *KeyNotPresent:
				if !node.tag.options.optional {
					errs = append(errs, err)
				}
			default:
				errs = append(errs, err)
			}
		}
		node.valueOf.Set(newVal)

		// Queue new nodes from it if needed
		frames, err := l.Queue(node.valueOf, node.tag)
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
	return errors.Join(errs...)
}

func MapLoader(data map[string]any) Loader {
	return func(dest any) (errs errorList) {
		// This function should be called with a struct
		stack := new(Stack)
		initFrame := &StackFrame{
			typeOf:  reflect.TypeOf(dest).Elem(),
			valueOf: reflect.ValueOf(dest).Elem(),
			data:    data,
			tag:     Tag{},
		}
		stack.Push(initFrame)

		for stack.Size() > 0 {
			// 	elem := stack.Pop()
			// 	switch elem.typeOf.Kind() {
			// 	case reflect.String:
			// 		strData, ok := elem.data.(string)
			// 		if !ok {
			// 			goto onError
			// 		}
			// 		elem.valueOf.SetString(strData)
			// 	case reflect.Int:
			// 		intData, ok := elem.data.(int)
			// 		if !ok {
			// 			goto onError
			// 		}
			// 		elem.valueOf.SetInt(int64(intData))
			// 	case reflect.Struct:
			// 		structData, ok := elem.data.(map[string]any)
			// 		if !ok {
			// 			goto onError
			// 		}
			// 		structVal := reflect.New(elem.typeOf).Elem()
			// 		elem.valueOf.Set(structVal)
			// 		queueStruct(stack, structData, elem)
			// 	case reflect.Slice:
			// 		sliceData, ok := elem.data.([]any)
			// 		if !ok {
			// 			goto onError
			// 		}
			// 		sliceVal := reflect.MakeSlice(elem.typeOf, len(sliceData), len(sliceData))
			// 		elem.valueOf.Set(sliceVal)
			// 		queueSlice(stack, sliceData, elem)
			// 	}
			// 	continue
			// onError:
			// 	if elem.data == nil {
			// 		if !elem.tag.options.optional {
			// 			errs = append(errs, errKeyNotPresent(elem.tag.key))
			// 		}
			// 	} else {
			// 		errs = append(errs, errInvalidValue(elem.tag.key))
			// 	}
		}

		return errs
	}
}

func FileLoader(configFile string, ignoreMissing bool) Loader {
	file, err := loadYamlFile(configFile)
	if err != nil {
		if ignoreMissing {
			return nilLoaderFn
		} else {
			panic(err)
		}
	}
	return MapLoader(file)
}

func EnvironmentLoader(envPrefix string) Loader {
	return func(dest any) (errs errorList) {
		// This function should be called with a struct
		stack := new(Stack)
		initFrame := &StackFrame{
			typeOf:  reflect.TypeOf(dest).Elem(),
			valueOf: reflect.ValueOf(dest).Elem(),
			tag:     Tag{},
		}
		stack.Push(initFrame)

		for stack.Size() > 0 {
			// 	elem := stack.Pop()
			// 	switch elem.typeOf.Kind() {
			// 	case reflect.String:
			// 		tagSegments := []string{}
			// 		tagSegments = append(
			// 			tagSegments,
			// 			strings.ToUpper(envPrefix),
			// 		)
			// 		for _, v := range elem.tag.path {
			// 			tagSegments = append(
			// 				tagSegments,
			// 				strings.ToUpper(v),
			// 			)
			// 		}
			// 		tagSegments = append(
			// 			tagSegments,
			// 			strings.ToUpper(elem.tag.key),
			// 		)
			// 		envName := strings.Join(tagSegments, "_")
			// 		envValue, ok := os.LookupEnv(envName)
			// 		fmt.Println("Checking env var", envName)
			// 		if !ok {
			// 			goto onError
			// 		}
			// 		elem.valueOf.SetString(envValue)
			// 	case reflect.Struct:
			// 		structVal := reflect.New(elem.typeOf).Elem()
			// 		elem.valueOf.Set(structVal)
			// 		queueStruct(stack, nil, elem)
			// 	}
			// 	continue
			// onError:
			// 	if elem.data == nil {
			// 		if !elem.tag.options.optional {
			// 			errs = append(errs, errKeyNotPresent(elem.tag.key))
			// 		}
			// 	} else {
			// 		errs = append(errs, errInvalidValue(elem.tag.key))
			// 	}
		}

		return errs
	}
}
