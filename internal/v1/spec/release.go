package spec

import (
	"flag"
)

const (
	defaultModel = "master"
	defaultBuild = false
)

type (
	// Release specifies the configurations for release command
	Release struct {
		Model string `json:"model" yaml:"model"`
		Build bool   `json:"build" yaml:"build"`
	}
)

// SetDefaults set default values for empty fields
func (r *Release) SetDefaults() {
	if r.Model == "" {
		r.Model = defaultModel
	}

	if r.Build == false {
		r.Build = defaultBuild
	}
}

// FlagSet returns a flag set for parsing input arguments
func (r *Release) FlagSet() *flag.FlagSet {
	fs := flag.NewFlagSet("release", flag.ContinueOnError)
	fs.StringVar(&r.Model, "model", r.Model, "")
	fs.BoolVar(&r.Build, "build", r.Build, "")

	return fs
}
