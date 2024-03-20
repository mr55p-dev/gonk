package gonk

import (
	"fmt"
	"reflect"
	"strings"
)

type tagOptions struct {
	optional bool
}

type tagData struct {
	config  string
	key     string
	path    []string
	options tagOptions
}

func parseTagOptions(opts []string) tagOptions {
	out := tagOptions{}
	for _, v := range opts {
		switch v {
		case "optional":
			out.optional = true
		}
	}
	return out
}

func parseConfigTag(config string) tagData {
	data := tagData{
		config: config,
	}
	v := strings.Split(config, ",")
	if len(v) > 1 {
		data.options = parseTagOptions(v[1:])
	}

	path := strings.Split(v[0], ".")
	data.key = path[len(path)-1]
	data.path = path[:len(path)-1]
	return data
}

func LoadConfig(dest any, loaders ...Loader) error {
	var err error
	defer func() {
		if msg := recover(); msg != nil {
			fmt.Println("Something paniced", msg)
			err = fmt.Errorf("Panic generated: %s", msg)
		}
	}()

	// for each loader, do some loading
	for idx, loader := range loaders {
		errs := make(map[string]error)
		applyLoader(loader, dest, errs)

		// handle errors
		for _, err := range errs {
			switch err.(type) {
			case *KeyNotPresent:
				if idx == len(loaders)-1 {
					return err
				}
			default:
				return err
			}
		}
		return nil
	}

	return err
}

type errors map[string]error

func tagPathConcat(original tagData, parts []string) tagData {
	out := original
	out.path = append(parts, out.path...)
	return out
}

func applyLoader(fn Loader, dest any, errs errors, prefix ...string) {
	valueOf := reflect.ValueOf(dest)
	for valueOf.Kind() == reflect.Pointer {
		valueOf = valueOf.Elem()
	}
	typeOf := valueOf.Type()

	if typeOf.Kind() != reflect.Struct {
		panic("Applying loader on non-struct type")
	}

	for i := 0; i < valueOf.NumField(); i++ {
		var err error
		fieldType := typeOf.Field(i)
		fieldValue := valueOf.Field(i)
		tagRaw, ok := fieldType.Tag.Lookup("config")
		if !ok {
			continue
		}

		tagParsed := parseConfigTag(tagRaw)
		tagParsed = tagPathConcat(tagParsed, prefix)
		switch fieldType.Type.Kind() {
		case reflect.Struct:
			newValuePtr := reflect.New(fieldType.Type)
			applyLoader(
				fn,
				newValuePtr.Interface(),
				errs,
				tagParsed.key,
			)
			fieldValue.Set(newValuePtr.Elem())
		case reflect.Array:
			
		case reflect.String, reflect.Int:
			err = fn(fieldType, fieldValue, tagParsed)
		default:

		}
		if err != nil {
			if _, ok := err.(*KeyNotPresent); ok && tagParsed.options.optional {
				continue
			} else {
				errs[tagParsed.config] = err
			}
		}
	}
}
