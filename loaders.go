package gonk

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
)

type Loader func(fieldType reflect.StructField, fieldValue reflect.Value, tag tagData) error

func nilLoaderFn(fieldType reflect.StructField, fieldValue reflect.Value, tag tagData) error {
	return nil
}

func MapLoader(data map[string]any) Loader {
	return func(fieldType reflect.StructField, fieldValue reflect.Value, tag tagData) error {
		// Set the value
		switch fieldType.Type.Kind() {
		case reflect.String:
			var value string
			err := traverseMap[string](&value, data, tag.key, tag.path...)
			if err != nil {
				return err
			}
			fieldValue.SetString(value)
		case reflect.Int:
			var value int
			err := traverseMap[int](&value, data, tag.key, tag.path...)
			if err != nil {
				return err
			}
			fieldValue.SetInt(int64(value))
		case reflect.Struct:
			structValue := reflect.Zero(fieldType.Type)
			structType := structValue.Type()
			structLoader := MapLoader(data)
			for i := 0; i < structValue.NumField(); i++ {
				fieldVal := structValue.Field(i)
				fieldType := structType.Field(i)
				fieldTag := fieldType.Tag.Get("config")
				if fieldTag == "" {
					fmt.Println("Skipping field")
					continue
				}
				fieldTagData := parseConfigTag(fieldTag)
				fmt.Printf("fieldTagData: %+v\n", fieldTagData)
				err := structLoader(fieldType, fieldVal, fieldTagData)
				if err != nil {
					fmt.Println("Error", err.Error())
					return err
				}
			}
			fieldValue.Set(structValue)
		case reflect.Array:
			return nil
		default:
			return fmt.Errorf("Invalid field type")
		}
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

func EnvironmentLoader(envPrefix string) Loader {
	return func(fieldType reflect.StructField, fieldValue reflect.Value, tag tagData) error {
		// Read the environment variables
		envName := getEnvName(tag.key, envPrefix)
		envVal, ok := os.LookupEnv(envName)
		if !ok {
			return &KeyNotPresent{"Key expected in variable " + envName}
		}
		switch fieldType.Type.Kind() {
		case reflect.String:
			fieldValue.SetString(envVal)
		case reflect.Int:
			envValInt, err := strconv.Atoi(envVal)
			if err != nil {
				return err
			}
			fieldValue.SetInt(int64(envValInt))
		}
		return nil
	}
}
