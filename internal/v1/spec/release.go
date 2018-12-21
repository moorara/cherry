package spec

import (
	"flag"
)

const (
	defaultBuild = false
)

type (
	// Release specifies the configurations for release command
	Release struct {
		Build bool `json:"build" yaml:"build"`
	}
)

// SetDefaults set default values for empty fields
func (r *Release) SetDefaults() {
	if r.Build == false {
		r.Build = defaultBuild
	}
}

// FlagSet returns a flag set for parsing input arguments
func (r *Release) FlagSet() *flag.FlagSet {
	fs := flag.NewFlagSet("release", flag.ContinueOnError)
	fs.BoolVar(&r.Build, "build", r.Build, "")

	return fs
}
