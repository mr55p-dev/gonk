package gonk

import (
	"errors"
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

func tagPathConcat(newTag tagData, prevKey string) tagData {
	out := newTag
	out.path = append(out.path, prevKey)
	return out
}
