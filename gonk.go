package gonk

import (
	"errors"
	"fmt"
	"strings"
)

type tagOptions struct {
	optional bool
}

type Tag struct {
	path    []any
	options tagOptions
}

func (t Tag) Key() any {
	return t.path[len(t.path)-1]
}

func (t Tag) String() string {
	arr := make([]string, len(t.path))
	for i := 0; i < len(arr); i++ {
		switch t.path[i].(type) {
		case string:
			arr[i] = t.path[i].(string)
		case int:
			arr[i] = fmt.Sprintf("[%d]", t.path[i].(int))
		}
	}
	return strings.Join(arr, ".")
}

func (t Tag) Push(component Tag) Tag {
	return Tag{
		path:    append(t.path, component.path...),
		options: component.options,
	}
}

func parseConfigTag(config string) Tag {
	data := Tag{}
	v := strings.Split(config, ",")

	if len(v) > 1 {
		for _, opt := range v[1:] {
			switch opt {
			case "optional":
				data.options.optional = true
			}
		}
	}

	path := strings.Split(v[0], ".")
	for _, str := range path {
		data.path = append(data.path, str)
	}
	return data
}

func LoadConfig(dest any, loaders ...Loader) error {
	// for each loader, do some loading
	validErrors := make(errorList, 0)
	for _, loader := range loaders {
		errs := loader(dest)
		for idx, err := range errs {
			switch err.(type) {
			case *KeyNotPresent:
				if idx == len(loaders)-1 {
					validErrors = append(validErrors, err)
				}
			default:
				validErrors = append(validErrors, err)
			}
		}
	}
	if len(validErrors) == 0 {
		return nil
	}

	return errors.Join(validErrors...)
}

type errorList []error

func tagPathConcat(newTag Tag, prevKey string) Tag {
	out := newTag
	out.path = append(out.path, prevKey)
	return out
}
