package spec

import (
	"flag"
)

const (
	defaultMainFile     = "main.go"
	defaultBinaryFile   = "bin/app"
	defaultCrossCompile = false
)

var (
	defaultGoVersions = []string{"1.11"}
	defaultPlatforms  = []string{"linux-386", "linux-amd64", "darwin-386", "darwin-amd64", "windows-386", "windows-amd64"}
)

type (
	// Build represents a build artifact
	Build struct {
		MainFile     string   `json:"mainFile" yaml:"main_file"`
		BinaryFile   string   `json:"binaryFile" yaml:"binary_file"`
		CrossCompile bool     `json:"crossCompile" yaml:"cross_compile"`
		GoVersions   []string `json:"goVersions" yaml:"go_versions"`
		Platforms    []string `json:"platforms" yaml:"platforms"`
	}
)

// SetDefaults set default values for empty fields
func (b *Build) SetDefaults() {
	if b.MainFile == "" {
		b.MainFile = defaultMainFile
	}

	if b.BinaryFile == "" {
		b.BinaryFile = defaultBinaryFile
	}

	if b.CrossCompile == false {
		b.CrossCompile = defaultCrossCompile
	}

	if b.GoVersions == nil || len(b.GoVersions) == 0 {
		b.GoVersions = defaultGoVersions
	}

	if b.Platforms == nil || len(b.Platforms) == 0 {
		b.Platforms = defaultPlatforms
	}
}

// FlagSet returns a flag set for parsing input arguments
func (b *Build) FlagSet() *flag.FlagSet {
	fs := flag.NewFlagSet("build", flag.ContinueOnError)
	fs.StringVar(&b.MainFile, "main-file", b.MainFile, "")
	fs.StringVar(&b.BinaryFile, "binary-file", b.BinaryFile, "")
	fs.BoolVar(&b.CrossCompile, "cross-compile", b.CrossCompile, "")

	return fs
}
