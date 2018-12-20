package spec

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type (
	// Spec is Cherry specs
	Spec struct {
		Version string  `json:"version" yaml:"version"`
		Builds  []Build `json:"build" yaml:"build"`
	}
)

// Read reads and returns a Spec from a YAML file
func Read(path string) (*Spec, error) {
	spec := new(Spec)
	path = filepath.Clean(path)

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	err = yaml.NewDecoder(f).Decode(spec)
	if err != nil {
		return nil, err
	}

	for i := range spec.Builds {
		spec.Builds[i] = spec.Builds[i].Defaults()
	}

	return spec, nil
}
