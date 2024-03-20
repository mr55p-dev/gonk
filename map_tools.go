package gonk

import (
	"os"

	"gopkg.in/yaml.v3"
)

func traverse(m any, tag Tag) (any, error) {
	// Traverse the config file
	head := m
	for _, component := range tag.path {
		switch component.(type) {
		case string:
			// head must be a map
			headMap, ok := head.(map[string]any)
			if !ok {
				return nil, errKeyNotPresent(tag.String())
			}
			head = headMap[component.(string)]
		case int:
			// head must be an array
			headSlice, ok := head.([]any)
			if !ok {
				return nil, errKeyNotPresent(tag.String())
			}
			head = headSlice[component.(int)]
		}
	}
	return head, nil
}

func loadYamlFile(filename string) (map[string]any, error) {
	out := make(map[string]any)
	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(file, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}
