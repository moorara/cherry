package spec

import (
	"encoding/json"
	"errors"
	"flag"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

const (
	defaultToolName       = "cherry"
	defaultVersion        = "1.0"
	defaultLanguage       = "go"
	defaultMainFile       = "main.go"
	defaultVersionPackage = "./version"
)

var (
	specFiles         = []string{"cherry.yml", "cherry.yaml", "cherry.json"}
	defaultGoVersions = []string{"1.15"}
	defaultPlatforms  = []string{"linux-386", "linux-amd64", "linux-arm", "linux-arm64", "darwin-386", "darwin-amd64", "windows-386", "windows-amd64"}
)

// Spec has all the specifications for Cherry.
type Spec struct {
	ToolName    string `json:"-" yaml:"-"`
	ToolVersion string `json:"-" yaml:"-"`

	Version  string  `json:"version" yaml:"version"`
	Language string  `json:"language" yaml:"language"`
	Build    Build   `json:"build" yaml:"build"`
	Release  Release `json:"release" yaml:"release"`
}

// FromFile reads and returns specifications from a file.
// If no spec file is found, a new spec with zero values will be returned.
func FromFile() (*Spec, error) {
	for _, file := range specFiles {
		ext := filepath.Ext(file)
		path := filepath.Clean(file)

		f, err := os.Open(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		defer f.Close()

		spec := new(Spec)

		if ext == ".yml" || ext == ".yaml" {
			err = yaml.NewDecoder(f).Decode(spec)
		} else if ext == ".json" {
			err = json.NewDecoder(f).Decode(spec)
		} else {
			return nil, errors.New("unknown spec file")
		}

		if err != nil {
			return nil, err
		}

		return spec, nil
	}

	return new(Spec), nil
}

// WithDefaults returns a new object with default values.
func (s Spec) WithDefaults() Spec {
	s.ToolName = "cherry"

	if s.Version == "" {
		s.Version = defaultVersion
	}

	if s.Language == "" {
		s.Language = defaultLanguage
	}

	s.Build = s.Build.WithDefaults()
	s.Release = s.Release.WithDefaults()

	return s
}

// Build has the specifications for build command.
type Build struct {
	CrossCompile   bool     `json:"crossCompile" yaml:"cross_compile"`
	MainFile       string   `json:"mainFile" yaml:"main_file"`
	BinaryFile     string   `json:"binaryFile" yaml:"binary_file"`
	VersionPackage string   `json:"versionPackage" yaml:"version_package"`
	GoVersions     []string `json:"goVersions" yaml:"go_versions"`
	Platforms      []string `json:"platforms" yaml:"platforms"`
}

// WithDefaults returns a new object with default values.
func (b Build) WithDefaults() Build {
	defaultBinaryFile := "bin/app"
	if wd, err := os.Getwd(); err == nil {
		defaultBinaryFile = "bin/" + filepath.Base(wd)
	}

	if b.MainFile == "" {
		b.MainFile = defaultMainFile
	}

	if b.BinaryFile == "" {
		b.BinaryFile = defaultBinaryFile
	}

	if b.VersionPackage == "" {
		b.VersionPackage = defaultVersionPackage
	}

	if len(b.GoVersions) == 0 {
		b.GoVersions = defaultGoVersions
	}

	if len(b.Platforms) == 0 {
		b.Platforms = defaultPlatforms
	}

	return b
}

// FlagSet returns a flag set for arguments of build command.
func (b *Build) FlagSet() *flag.FlagSet {
	fs := flag.NewFlagSet("build", flag.ContinueOnError)
	fs.BoolVar(&b.CrossCompile, "cross-compile", b.CrossCompile, "")
	fs.StringVar(&b.MainFile, "main-file", b.MainFile, "")
	fs.StringVar(&b.BinaryFile, "binary-file", b.BinaryFile, "")
	fs.StringVar(&b.VersionPackage, "version-package", b.VersionPackage, "")

	return fs
}

// Release has the specifications for release command.
type Release struct {
	Build bool `json:"build" yaml:"build"`
}

// WithDefaults returns a new object with default values.
func (r Release) WithDefaults() Release {
	return r
}

// FlagSet returns a flag set for arguments of release command.
func (r *Release) FlagSet() *flag.FlagSet {
	fs := flag.NewFlagSet("release", flag.ContinueOnError)
	fs.BoolVar(&r.Build, "build", r.Build, "")

	return fs
}
