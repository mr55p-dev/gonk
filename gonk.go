package gonk

import (
	"fmt"
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

func applyLoader(fn Loader, dest any, errs errors) {
	fmt.Println("Starting calling loader")
	_ = fn(dest)
	fmt.Printf("errs: %v\n", errs)
	// if err != nil {
	// 	if _, ok := err.(*KeyNotPresent); ok && tagParsed.options.optional {
	// 		continue
	// 	} else {
	// 		errs[tagParsed.config] = err
	// 	}
	// }
}
