package gonk

import (
	"fmt"
	"os"
	"reflect"
	"strings"
)

type Loader func(dest any) errorList

func nilLoaderFn(dest any) errorList {
	return nil
}

type StackFrame struct {
	valueOf reflect.Value
	typeOf  reflect.Type
	tag     tagData
	data    any
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

func queueStruct(stack *Stack, data map[string]any, baseFrame *StackFrame) {
	for i := 0; i < baseFrame.typeOf.NumField(); i++ {
		// Create a new stack frame
		frame := new(StackFrame)
		frame.typeOf = baseFrame.typeOf.Field(i).Type
		frame.valueOf = baseFrame.valueOf.Field(i)
		// Parse the field tag
		tagRaw, ok := baseFrame.typeOf.Field(i).Tag.Lookup("config")
		if !ok {
			continue
		}
		frame.tag = parseConfigTag(tagRaw)
		if baseFrame.tag.key != "" {
			frame.tag = tagPathConcat(frame.tag, baseFrame.tag.key)
		}
		frame.data = data[frame.tag.key]
		stack.Push(frame)
	}
	return
}

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
	Init() error
	SetFromNode(node reflect.Value) (reflect.Value, error)
	Queue(node reflect.Value) ([]*StackFrame, error)
	CanQueue(node reflect.Type) bool
	CanRead(node reflect.Type) bool
}

func GenericLoader(target any, l loader) {
	nodeStk := new(Stack)
	frames, err := l.Queue(reflect.ValueOf(target).Elem())
	for nodeStk.Size() > 0 {
		elem := nodeStk.Pop()
		if l.CanQueue(elem.typeOf) {
			frames, err := l.Queue(nodeStk)
			if err != nil {
				panic(err)
			}
			for _, frame := range frames {
				nodeStk.Push(frame)
			}
		}
		if l.CanRead(elem.typeOf) {
			l.SetFromNode(elem.valueOf, elem.tag)
		}
	}
}

/*
- traverse the struct
- for each field of the struct, tagged with config
	- parse the field tag
	- check the node type
	- case string/int/settable:
		- run setter
	- case struct/arr:
		- traverse nested struct
*/

func MapLoader(data map[string]any) Loader {
	return func(dest any) (errs errorList) {
		// This function should be called with a struct
		stack := new(Stack)
		initFrame := &StackFrame{
			typeOf:  reflect.TypeOf(dest).Elem(),
			valueOf: reflect.ValueOf(dest).Elem(),
			data:    data,
			tag:     tagData{},
		}
		stack.Push(initFrame)

		for stack.Size() > 0 {
			elem := stack.Pop()
			switch elem.typeOf.Kind() {
			case reflect.String:
				strData, ok := elem.data.(string)
				if !ok {
					goto onError
				}
				elem.valueOf.SetString(strData)
			case reflect.Int:
				intData, ok := elem.data.(int)
				if !ok {
					goto onError
				}
				elem.valueOf.SetInt(int64(intData))
			case reflect.Struct:
				structData, ok := elem.data.(map[string]any)
				if !ok {
					goto onError
				}
				structVal := reflect.New(elem.typeOf).Elem()
				elem.valueOf.Set(structVal)
				queueStruct(stack, structData, elem)
			case reflect.Slice:
				sliceData, ok := elem.data.([]any)
				if !ok {
					goto onError
				}
				sliceVal := reflect.MakeSlice(elem.typeOf, len(sliceData), len(sliceData))
				elem.valueOf.Set(sliceVal)
				queueSlice(stack, sliceData, elem)
			}
			continue
		onError:
			if elem.data == nil {
				if !elem.tag.options.optional {
					errs = append(errs, errKeyNotPresent(elem.tag.key))
				}
			} else {
				errs = append(errs, errInvalidValue(elem.tag.key))
			}
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
			tag:     tagData{},
		}
		stack.Push(initFrame)

		for stack.Size() > 0 {
			elem := stack.Pop()
			switch elem.typeOf.Kind() {
			case reflect.String:
				tagSegments := []string{}
				tagSegments = append(
					tagSegments,
					strings.ToUpper(envPrefix),
				)
				for _, v := range elem.tag.path {
					tagSegments = append(
						tagSegments,
						strings.ToUpper(v),
					)
				}
				tagSegments = append(
					tagSegments,
					strings.ToUpper(elem.tag.key),
				)
				envName := strings.Join(tagSegments, "_")
				envValue, ok := os.LookupEnv(envName)
				fmt.Println("Checking env var", envName)
				if !ok {
					goto onError
				}
				elem.valueOf.SetString(envValue)
			case reflect.Struct:
				structVal := reflect.New(elem.typeOf).Elem()
				elem.valueOf.Set(structVal)
				queueStruct(stack, nil, elem)
			}
			continue
		onError:
			if elem.data == nil {
				if !elem.tag.options.optional {
					errs = append(errs, errKeyNotPresent(elem.tag.key))
				}
			} else {
				errs = append(errs, errInvalidValue(elem.tag.key))
			}
		}

		return errs
	}
}
