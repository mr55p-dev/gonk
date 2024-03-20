package gonk

import (
	"fmt"
	"reflect"
)

type Loader func(dest any) errors

func nilLoaderFn(dest any) errors {
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

func queueStruct(stack *Stack, data map[string]any, errs errors, baseFrame *StackFrame) {
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
		if !ok {
			errs[frame.valueOf.Type().Name()] = errInvalidValue("")
		}
		frame.data = data[frame.tag.key]
		stack.Push(frame)
	}
}

func MapLoader(data map[string]any) Loader {
	return func(dest any) errors {
		// This function should be called with a struct
		errs := make(errors)
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
					errs[elem.tag.key] = errInvalidValue(elem.tag.key)
					continue
				}
				elem.valueOf.SetString(strData)
				fmt.Printf("Setting %s to %s\n", elem.typeOf.Name(), strData)
			case reflect.Int:
				intData, ok := elem.data.(int)
				if !ok {
					errs[elem.tag.key] = errInvalidValue(elem.tag.key)
					continue
				}
				elem.valueOf.SetInt(int64(intData))
			case reflect.Struct:
				structData, ok := elem.data.(map[string]any)
				if !ok {
					errs[elem.tag.key] = errInvalidValue(elem.tag.key)
				}
				structVal := reflect.New(elem.typeOf).Elem()
				elem.valueOf.Set(structVal)
				queueStruct(stack, structData, errs, elem)
			case reflect.Slice:
				arrData, ok := elem.data.([]any)
				if !ok {
					errs[elem.tag.key] = errInvalidValue(elem.tag.key)
					continue
				}
				// Allocate a new array of the type
				newSlice := reflect.MakeSlice(
					elem.typeOf,
					len(arrData),
					len(arrData),
				)
				elem.valueOf.Set(newSlice)
				for idx, arrElem := range arrData {
					newFrame := new(StackFrame)
					newFrame.data = arrElem
					newFrame.typeOf = elem.typeOf.Elem()
					newFrame.valueOf = newSlice.Index(idx)

					stack.Push(newFrame)
				}
				fmt.Printf("newSlice.Interface(): %v\n", newSlice.Interface())
			}
		}

		fmt.Printf("dest: %#v\n", dest)
		fmt.Printf("errs: %#v\n", errs)
		fmt.Printf("initFrame.valueOf.Interface(): %v\n", initFrame.valueOf.Interface())

		return nil
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
