package model

import (
	"runtime"
	"strings"
)

type (
	// Build is the build specs
	Build struct {
		Language string   `json:"language" yaml:"language"`
		Version  []string `json:"version" yaml:"version"`
		OS       []string `json:"os" yaml:"os"`
		Arch     []string `json:"arch" yaml:"arch"`
	}
)

// Defaults returns a new build with default values
func (b Build) Defaults() Build {
	var defaultVersion, defaultOS, defaultArch string

	switch strings.ToLower(b.Language) {
	case "go":
		defaultVersion = "1.11.2"
		defaultOS = runtime.GOOS
		defaultArch = runtime.GOARCH
	}

	if b.Version == nil || len(b.Version) == 0 {
		if defaultVersion == "" {
			b.Version = []string{}
		} else {
			b.Version = []string{defaultVersion}
		}
	}

	if b.OS == nil || len(b.OS) == 0 {
		if defaultOS == "" {
			b.OS = []string{}
		} else {
			b.OS = []string{defaultOS}
		}
	}

	if b.Arch == nil || len(b.Arch) == 0 {
		if defaultArch == "" {
			b.Arch = []string{}
		} else {
			b.Arch = []string{defaultArch}
		}
	}

	return b
}
