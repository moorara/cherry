package spec

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildSetDefaults(t *testing.T) {
	tests := []struct {
		build         Build
		expectedBuild Build
	}{
		{
			Build{},
			Build{
				MainFile:     defaultMainFile,
				BinaryFile:   defaultBinaryFile,
				CrossCompile: defaultCrossCompile,
				GoVersions:   defaultGoVersions,
				Platforms:    defaultPlatforms,
			},
		},
		{
			Build{
				MainFile:     "cmd/main.go",
				BinaryFile:   "build/app",
				CrossCompile: true,
				GoVersions:   []string{"1.10", "1.11"},
				Platforms:    []string{"linux-amd64", "darwin-amd64", "windows-amd64"},
			},
			Build{
				MainFile:     "cmd/main.go",
				BinaryFile:   "build/app",
				CrossCompile: true,
				GoVersions:   []string{"1.10", "1.11"},
				Platforms:    []string{"linux-amd64", "darwin-amd64", "windows-amd64"},
			},
		},
	}

	for _, tc := range tests {
		tc.build.SetDefaults()
		assert.NotNil(t, tc.expectedBuild, tc.build)
	}
}

func TestBuildFlagSet(t *testing.T) {
	tests := []struct {
		build        Build
		expectedName string
	}{
		{
			build:        Build{},
			expectedName: "build",
		},
		{
			build: Build{
				MainFile:     "main.go",
				BinaryFile:   "bin/app",
				CrossCompile: true,
				GoVersions:   []string{"1.10", "1.11"},
				Platforms:    []string{"linux-386", "linux-amd64", "darwin-386", "darwin-amd64", "windows-386", "windows-amd64"},
			},
			expectedName: "build",
		},
	}

	for _, tc := range tests {
		fs := tc.build.FlagSet()
		assert.Equal(t, tc.expectedName, fs.Name())
	}
}
