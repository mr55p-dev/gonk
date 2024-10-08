// Package gonk is a configuration loader built using reflection.
// Given a public struct with public fields tagged `config`, gonk will load values from the given
// sources in the order specified. a struct pointer tagged with `config:"name"` fields or a slice,
// [LoadConfig] can be used to load values from the [Loader]s into the struct.
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
// In this example, gonk will look for environment variables named `PREFIX_1` and
// `PREFIX_ANOTHERVAR`. If one or more values are not available in at least one source, gonk will
// return an error, unless the field is marked with the optional tag.
//
// gonk supports loading nested structs as well, provided they are also publicly exported
//
// Field tags may be more complex. The syntax is:
//
//	pattern: { [ path components ]... tag [ ,options... ] }
//
//	path components: allows for defining nested keys by specifying the path components separated with a '.'.
//
//	tag: name of the key to lookup by the loader. Loaders may concatenate previous path components into the name they
//	look up (ie environment loader will prefix all path components, upper cased and separated with an '_').
//
//	options: optional	specifies if the key can be omitted by a specific loader.
//
// For example:
//
//	type Example struct {
//		FirstField   int    `config:"app.my-int"`
//		SecondField  string `config:"my-string"`
//		NestedStruct struct {
//			AnotherField int `config:"my-int"` // Will load `app.my-int`
//		} `config:"app"`
//	}
package gonk

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"unicode"
)

type errorList []error

type Loader interface {
	Load(node reflect.Value, tag tagData) (reflect.Value, error)
}

type loadState int

const (
	loadRequired loadState = iota
	loadOptional
	loadComplete
)

// LoadConfig loads configuration into a struct pointer or slice. Pass one or more loaders as arguments to provide
// sources to load from. Loaders will be applied in the order given here. Produces a joined error containing all errors
// encountered whilst loading, or nil if the loading was succesfull. Missing fields marked as optional will not be
// reported as errors.
func LoadConfig(dest any, loaders ...Loader) error {
	loaded := make(map[string]loadState)
	val := reflect.ValueOf(dest)
	missingValues := make([]error, 0)

	// Setup queue of nodes to fetch and init with all the values passed in
	nodeStk := new(stack)
	frames, err := queueNode(val, tagData{})
	if err != nil {
		return err
	}
	nodeStk.push(frames...)

	// Empty the queue of nodes
	for nodeStk.size() > 0 {
		node := nodeStk.pop()
		nodeId := node.tag.String()
		if node.tag.options.optional {
			loaded[nodeId] = loadOptional
		} else {
			loaded[nodeId] = loadRequired
		}

		// Try to load the value
		err := applyLoaders(node, nodeId, loaded, loaders...)
		if err != nil {
			return err
		}

		// Check to see if any of the loaders succeeded, but only if the value is required
		if loaded[nodeId] == loadRequired {
			missingValues = append(missingValues, ValueNotPresentError(
				fmt.Sprintf("No value in any loader for %s", nodeId),
			))
		}

		// Queue up the next set of nodes
		frames, err := queueNode(node.valueOf, node.tag)
		if err != nil {
			return err
		}
		nodeStk.push(frames...)
	}
	if len(missingValues) > 0 {
		return errors.Join(missingValues...)
	}
	return nil
}

func applyLoaders(node *nodeFrame, nodeId string, loaded map[string]loadState, loaders ...Loader) error {
	for _, loader := range loaders {
		if loader == nil {
			continue
		}
		res, err := loader.Load(node.valueOf, node.tag)
		if err != nil {
			// exit on errors, except value not present, which we skip
			if err, ok := err.(ValueNotPresentError); !ok {
				return err
			} else {
				continue
			}
		}
		node.valueOf.Set(res)
		loaded[nodeId] = loadComplete
	}
	return nil
}

func kebabCase(s string) string {
	output := new(strings.Builder)
	for idx, ch := range s {
		if unicode.IsUpper(ch) {
			if idx > 0 {
				output.WriteRune('-')
			}
			output.WriteRune(unicode.ToLower(ch))
		} else {
			output.WriteRune(ch)
		}
	}
	return output.String()
}

func queueNode(node reflect.Value, tag tagData) (out []*nodeFrame, err error) {
	switch node.Kind() {
	case reflect.Struct:
		nodeType := node.Type()
		for i := 0; i < nodeType.NumField(); i++ {
			newFrame := new(nodeFrame)
			fieldType := nodeType.Field(i)
			tagRaw, ok := fieldType.Tag.Lookup("config")
			if ok && tagRaw == "-" {
				continue
			} else if !ok {
				tagRaw = kebabCase(fieldType.Name)
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
