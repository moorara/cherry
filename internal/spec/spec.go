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
	defaultCrossCompile   = false
	defaultMainFile       = "main.go"
	defaultVersionPackage = "./cmd/version"

	defaultModel = "master"
	defaultBuild = false

	defaultCoverMode  = "atomic"
	defaultReportPath = "coverage"

	defaultToolName    = "cherry"
	defaultVersion     = "1.0"
	defaultLanguage    = "go"
	defaultVersionFile = "VERSION"
)

var (
	specFiles = []string{"cherry.yml", "cherry.yaml", "cherry.json"}

	defaultGoVersions = []string{"1.13"}
	defaultPlatforms  = []string{"linux-386", "linux-amd64", "linux-arm", "linux-arm64", "darwin-386", "darwin-amd64", "windows-386", "windows-amd64"}
)

// Build has the specifications for build command.
type Build struct {
	CrossCompile   bool     `json:"crossCompile" yaml:"cross_compile"`
	MainFile       string   `json:"mainFile" yaml:"main_file"`
	BinaryFile     string   `json:"binaryFile" yaml:"binary_file"`
	VersionPackage string   `json:"versionPackage" yaml:"version_package"`
	GoVersions     []string `json:"goVersions" yaml:"go_versions"`
	Platforms      []string `json:"platforms" yaml:"platforms"`
}

// SetDefaults sets default values for empty fields.
func (b *Build) SetDefaults() {
	defaultBinaryFile := "bin/app"
	if wd, err := os.Getwd(); err == nil {
		defaultBinaryFile = "bin/" + filepath.Base(wd)
	}

	if b.CrossCompile == false {
		b.CrossCompile = defaultCrossCompile
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
}

// FlagSet returns a flag set for input arguments for build command.
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
	Model string `json:"model" yaml:"model"`
	Build bool   `json:"build" yaml:"build"`
}

// SetDefaults sets default values for empty fields.
func (r *Release) SetDefaults() {
	if r.Model == "" {
		r.Model = defaultModel
	}

	if r.Build == false {
		r.Build = defaultBuild
	}
}

// FlagSet returns a flag set for input arguments for release command.
func (r *Release) FlagSet() *flag.FlagSet {
	fs := flag.NewFlagSet("release", flag.ContinueOnError)
	fs.StringVar(&r.Model, "model", r.Model, "")
	fs.BoolVar(&r.Build, "build", r.Build, "")

	return fs
}

// Test has the specifications for test command.
type Test struct {
	CoverMode  string `json:"coverMode" yaml:"cover_mode"`
	ReportPath string `json:"reportPath" yaml:"report_path"`
}

// SetDefaults sets default values for empty fields.
func (t *Test) SetDefaults() {
	if t.CoverMode == "" {
		t.CoverMode = defaultCoverMode
	}

	if t.ReportPath == "" {
		t.ReportPath = defaultReportPath
	}
}

// FlagSet returns a flag set for input arguments for test command.
func (t *Test) FlagSet() *flag.FlagSet {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.StringVar(&t.CoverMode, "covermode", t.CoverMode, "")
	fs.StringVar(&t.ReportPath, "report-path", t.ReportPath, "")

	return fs
}

// Spec has all the specifications for Cherry.
type Spec struct {
	ToolName    string `json:"-" yaml:"-"`
	ToolVersion string `json:"-" yaml:"-"`

	Version     string  `json:"version" yaml:"version"`
	Language    string  `json:"language" yaml:"language"`
	VersionFile string  `json:"versionFile" yaml:"version_file"`
	Test        Test    `json:"test" yaml:"test"`
	Build       Build   `json:"build" yaml:"build"`
	Release     Release `json:"release" yaml:"release"`
}

// SetDefaults sets default values for empty fields.
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

// ReadSpec reads and returns specifications from a file.
func ReadSpec() (*Spec, error) {
	for _, file := range specFiles {
		ext := filepath.Ext(file)
		path := filepath.Clean(file)

		f, err := os.Open(path)
		if err != nil {
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

	return nil, errors.New("no spec file found")
}
