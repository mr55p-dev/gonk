package gonk

import (
	"os"

	"gopkg.in/yaml.v3"
)

func traverseMap(target map[string]any, key string, segments ...string) (any, error) {
	// Traverse the config file
	configFileKey := target
	for _, segment := range segments {
		var ok bool
		configFileKey, ok = configFileKey[segment].(map[string]any)
		if !ok {
			return nil, errKeyNotPresent(key)
		}
	}
	value, ok := configFileKey[key]
	if !ok {
		return nil, errKeyNotPresent(key)
	}
	return value, nil
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
