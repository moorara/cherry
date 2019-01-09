package spec

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

const (
	// SpecFile is the default name of specification/configuration file
	SpecFile = "cherry.yaml"

	defaultToolName    = "cherry"
	defaultVersion     = "1.0"
	defaultLanguage    = "go"
	defaultVersionFile = "VERSION"
)

type (
	// Spec is Cherry specs
	Spec struct {
		ToolName    string `json:"-" yaml:"-"`
		ToolVersion string `json:"-" yaml:"-"`

		Version     string  `json:"version" yaml:"version"`
		Language    string  `json:"language" yaml:"language"`
		VersionFile string  `json:"versionFile" yaml:"version_file"`
		Test        Test    `json:"test" yaml:"test"`
		Build       Build   `json:"build" yaml:"build"`
		Release     Release `json:"release" yaml:"release"`
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

	return spec, nil
}

// SetDefaults set default values for empty fields
func (s *Spec) SetDefaults() {
	if s.ToolName == "" {
		s.ToolName = defaultToolName
	}

	if s.Version == "" {
		s.Version = defaultVersion
	}

	if s.Language == "" {
		s.Language = defaultLanguage
	}

	if s.VersionFile == "" {
		s.VersionFile = defaultVersionFile
	}

	s.Test.SetDefaults()
	s.Build.SetDefaults()
	s.Release.SetDefaults()
}
