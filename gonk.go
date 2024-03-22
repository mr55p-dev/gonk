// Package gonk is a configuration loader
// It can be used to import configuration from yaml files and environment variables currently. Given a struct pointer
// tagged with `config:"name"` fields or a slice, [LoadConfig] can be used to load values from the [Loader]s into the
// struct.
//
//	type Example struct {
//		Field1 string `config:"1"`
//	    Field2 string `config:"anotherVar"`
//	}
//
//	func main() {
//	    config :=new(Example)
//	    err := LoadConfig(config, EnvLoader("prefix"))
//	}
//
// In this example, gonk will look for environment variables named `PREFIX_1` and `PREFIX_ANOTHERVAR`. Field tags may be
// more complex. The syntax is:
//
//	pattern: { [ path components ]... tag [ ,options ] }
//
//	path components: allows for defining nested keys by specifying the path components separated with a '.'.
//
//	tag: name of the key to lookup by the loader. Loaders may concatenate previous path components into the name they
//	look up (ie environment loader will prefix all path components, upper cased and separated with an '_').
//
//	options: optional	specifies if the key can be omitted by a specific loader.
//
// When specifying multiple loaders, a key must be present in at least one loader in order for the config to be loaded
// succesfully, unless the field is marked as optional.
package gonk

import (
	"errors"
	"reflect"
)

type errorList []error

type Loader interface {
	Load(node reflect.Value, tag tagData) error
}

// LoadConfig loads configuration into a struct pointer or slice. Pass one or more loaders as arguments to provide
// sources to load from. Loaders will be applied in the order given here. Produces a joined error containing all errors
// encountered whilst loading, or nil if the loading was succesfull. Missing fields marked as optional will not be
// reported as errors.
func LoadConfig(dest any, loaders ...Loader) error {
	// for each loader, do some loading
	validErrors := make(errorList, 0)
	for idx, ldr := range loaders {
		errs := applyLoader(dest, ldr)
		for _, err := range errs {
			switch err.(type) {
			case *ValueNotPresent:
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

func queueNode(node reflect.Value, tag tagData) (out []*nodeFrame, err error) {
	switch node.Kind() {
	case reflect.Struct:
		nodeType := node.Type()
		for i := 0; i < nodeType.NumField(); i++ {
			newFrame := new(nodeFrame)
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
			frame := new(nodeFrame)
			frame.valueOf = node.Index(i)
			frame.tag = tag.Push(tagData{
				path: []any{i},
			})
			out = append(out, frame)
		}
		return
	case reflect.Pointer:
		return queueNode(node.Elem(), tag)
	default:
		return
	}
}

func applyLoader(target any, l Loader) errorList {
	errs := make(errorList, 0)
	nodeStk := new(stack)
	frames, err := queueNode(reflect.ValueOf(target), tagData{})
	if err != nil {
		errs = append(errs, err)
		return errs
	}
	for _, v := range frames {
		nodeStk.push(v)
	}
	for nodeStk.size() > 0 {
		node := nodeStk.pop()
		// Set the nodes value
		err := l.Load(node.valueOf, node.tag)
		if err != nil {
			switch err.(type) {
			case ValueNotPresent:
				if !node.tag.options.optional {
					errs = append(errs, err)
				}
			default:
				errs = append(errs, err)
			}
		}

		// Queue new nodes from it if needed
		frames, err := queueNode(node.valueOf, node.tag)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		for _, frame := range frames {
			nodeStk.push(frame)
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return errs
}
