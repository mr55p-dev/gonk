package gonk

import (
	"fmt"
	"reflect"
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
		fmt.Printf("data: %+v\n", data)

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

// func EnvironmentLoader(envPrefix string) Loader {
// 	return func(fieldType reflect.StructField, fieldValue reflect.Value, tag tagData) error {
// 		// Read the environment variables
// 		envName := getEnvName(tag.key, envPrefix)
// 		envVal, ok := os.LookupEnv(envName)
// 		if !ok {
// 			return &KeyNotPresent{"Key expected in variable " + envName}
// 		}
// 		switch fieldType.Type.Kind() {
// 		case reflect.String:
// 			fieldValue.SetString(envVal)
// 		case reflect.Int:
// 			envValInt, err := strconv.Atoi(envVal)
// 			if err != nil {
// 				return err
// 			}
// 			fieldValue.SetInt(int64(envValInt))
// 		}
// 		return nil
// 	}
// }
