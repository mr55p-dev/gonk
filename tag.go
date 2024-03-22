package gonk

import (
	"fmt"
	"strings"
)

type tagOptions struct {
	optional bool
}

type tagData struct {
	path    []any
	options tagOptions
}

// Key returns the topmost element of the path
func (t tagData) Key() any {
	if len(t.path) == 0 {
		return nil
	}
	return t.path[len(t.path)-1]
}

// String returns a string representation of the entire path
func (t tagData) String() string {
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

// Push merges the path of two tags, adding the argument as a path on top of the existing one. Does not update any existing tags
func (t tagData) Push(component tagData) tagData {
	return tagData{
		path:    append(t.path, component.path...),
		options: component.options,
	}
}

// NamedKeys returns only the object components of the path
func (t tagData) NamedKeys() []string {
	out := make([]string, 0)
	for _, val := range t.path {
		v, ok := val.(string)
		if ok {
			out = append(out, v)
		}
	}
	return out
}

func parseConfigTag(config string) tagData {
	data := tagData{}
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

func tagPathConcat(newTag tagData, prevKey string) tagData {
	out := newTag
	out.path = append(out.path, prevKey)
	return out
}
