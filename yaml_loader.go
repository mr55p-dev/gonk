package gonk

import (
	"os"

	"gopkg.in/yaml.v3"
)

// NewYamlLoader returns a loader for the yaml file specified in the path. If the file does not
// exist or cannot be unmarshalled, this returns an error.
func NewYamlLoader(file string) (Loader, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	mp := make(map[string]any)
	err = yaml.Unmarshal(data, mp)
	if err != nil {
		return nil, err
	}
	return MapLoader(mp), nil
}
