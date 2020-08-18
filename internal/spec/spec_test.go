package spec

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFromFile(t *testing.T) {
	tests := []struct {
		name          string
		specFiles     []string
		expectedSpec  Spec
		expectedError string
	}{
		{
			name:         "NoSpecFile",
			specFiles:    []string{"test/null"},
			expectedSpec: Spec{},
		},
		{
			name:          "UnknownFile",
			specFiles:     []string{"test/unknown"},
			expectedError: "unknown spec file",
		},
		{
			name:          "EmptyYAML",
			specFiles:     []string{"test/empty.yaml"},
			expectedError: "EOF",
		},
		{
			name:          "EmptyJSON",
			specFiles:     []string{"test/empty.json"},
			expectedError: "EOF",
		},
		{
			name:          "InvalidYAML",
			specFiles:     []string{"test/invalid.yaml"},
			expectedError: "cannot unmarshal",
		},
		{
			name:          "InvalidJSON",
			specFiles:     []string{"test/invalid.json"},
			expectedError: "invalid character",
		},
		{
			name:      "MinimumYAML",
			specFiles: []string{"test/min.yaml"},
			expectedSpec: Spec{
				Version:  "1.0",
				Language: "go",
				Build:    Build{},
				Release: Release{
					Build: true,
				},
			},
		},
		{
			name:      "MinimumJSON",
			specFiles: []string{"test/min.json"},
			expectedSpec: Spec{
				Version:  "1.0",
				Language: "go",
				Build:    Build{},
				Release: Release{
					Build: true,
				},
			},
		},
		{
			name:      "MaximumYAML",
			specFiles: []string{"test/max.yaml"},
			expectedSpec: Spec{
				Version:  "1.0",
				Language: "go",
				Build: Build{
					CrossCompile:   true,
					MainFile:       "main.go",
					BinaryFile:     "bin/cherry",
					VersionPackage: "./version",
					GoVersions:     []string{"1.15", "1.14.6", "1.12.x"},
					Platforms:      []string{"linux-386", "linux-amd64", "linux-arm", "linux-arm64", "darwin-386", "darwin-amd64", "windows-386", "windows-amd64"},
				},
				Release: Release{
					Build: true,
				},
			},
		},
		{
			name:      "MaximumJSON",
			specFiles: []string{"test/max.json"},
			expectedSpec: Spec{
				Version:  "1.0",
				Language: "go",
				Build: Build{
					CrossCompile:   true,
					MainFile:       "main.go",
					BinaryFile:     "bin/cherry",
					VersionPackage: "./version",
					GoVersions:     []string{"1.15", "1.14.6", "1.12.x"},
					Platforms:      []string{"linux-386", "linux-amd64", "linux-arm", "linux-arm64", "darwin-386", "darwin-amd64", "windows-386", "windows-amd64"},
				},
				Release: Release{
					Build: true,
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			specFiles = tc.specFiles
			spec, err := FromFile()

			if tc.expectedError != "" {
				assert.Contains(t, err.Error(), tc.expectedError)
				assert.Equal(t, Spec{}, spec)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedSpec, spec)
			}
		})
	}
}

func TestSpecWithDefaults(t *testing.T) {
	tests := []struct {
		spec         Spec
		expectedSpec Spec
	}{
		{
			Spec{},
			Spec{
				ToolName:    defaultToolName,
				ToolVersion: "",
				Version:     defaultVersion,
				Language:    defaultLanguage,
				Build: Build{
					CrossCompile:   false,
					MainFile:       defaultMainFile,
					BinaryFile:     "bin/spec",
					VersionPackage: defaultVersionPackage,
					GoVersions:     defaultGoVersions,
					Platforms:      defaultPlatforms,
				},
				Release: Release{
					Build: false,
				},
			},
		},
		{
			Spec{
				Version:  "2.0",
				Language: "go",
				Build: Build{
					CrossCompile:   true,
					MainFile:       "cmd/my-app/main.go",
					BinaryFile:     "build/my-app",
					VersionPackage: "./version",
					GoVersions:     []string{"1.15", "1.14.6"},
					Platforms:      []string{"linux-amd64", "darwin-amd64", "windows-amd64"},
				},
				Release: Release{
					Build: true,
				},
			},
			Spec{
				ToolName:    defaultToolName,
				ToolVersion: "",
				Version:     "2.0",
				Language:    "go",
				Build: Build{
					CrossCompile:   true,
					MainFile:       "cmd/my-app/main.go",
					BinaryFile:     "build/my-app",
					VersionPackage: "./version",
					GoVersions:     []string{"1.15", "1.14.6"},
					Platforms:      []string{"linux-amd64", "darwin-amd64", "windows-amd64"},
				},
				Release: Release{
					Build: true,
				},
			},
		},
	}

	for _, tc := range tests {
		assert.Equal(t, tc.expectedSpec, tc.spec.WithDefaults())
	}
}

func TestBuildWithDefaults(t *testing.T) {
	tests := []struct {
		build         Build
		expectedBuild Build
	}{
		{
			Build{},
			Build{
				CrossCompile:   false,
				MainFile:       defaultMainFile,
				BinaryFile:     "bin/spec",
				VersionPackage: defaultVersionPackage,
				GoVersions:     defaultGoVersions,
				Platforms:      defaultPlatforms,
			},
		},
		{
			Build{
				CrossCompile:   true,
				MainFile:       "cmd/my-app/main.go",
				BinaryFile:     "build/my-app",
				VersionPackage: "./version",
				GoVersions:     []string{"1.15", "1.14.6"},
				Platforms:      []string{"linux-amd64", "darwin-amd64", "windows-amd64"},
			},
			Build{
				CrossCompile:   true,
				MainFile:       "cmd/my-app/main.go",
				BinaryFile:     "build/my-app",
				VersionPackage: "./version",
				GoVersions:     []string{"1.15", "1.14.6"},
				Platforms:      []string{"linux-amd64", "darwin-amd64", "windows-amd64"},
			},
		},
	}

	for _, tc := range tests {
		assert.Equal(t, tc.expectedBuild, tc.build.WithDefaults())
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
				CrossCompile:   true,
				MainFile:       "main.go",
				BinaryFile:     "bin/app",
				VersionPackage: "./version",
				GoVersions:     []string{"1.15"},
				Platforms:      []string{"linux-386", "linux-amd64", "darwin-386", "darwin-amd64", "windows-386", "windows-amd64"},
			},
			expectedName: "build",
		},
	}

	for _, tc := range tests {
		fs := tc.build.FlagSet()
		assert.Equal(t, tc.expectedName, fs.Name())
	}
}

func TestReleaseWithDefaults(t *testing.T) {
	tests := []struct {
		release         Release
		expectedRelease Release
	}{
		{
			Release{},
			Release{
				Build: false,
			},
		},
		{
			Release{
				Build: true,
			},
			Release{
				Build: true,
			},
		},
	}

	for _, tc := range tests {
		assert.Equal(t, tc.expectedRelease, tc.release.WithDefaults())
	}
}

func TestReleaseFlagSet(t *testing.T) {
	tests := []struct {
		release      Release
		expectedName string
	}{
		{
			release:      Release{},
			expectedName: "release",
		},
		{
			release: Release{
				Build: true,
			},
			expectedName: "release",
		},
	}

	for _, tc := range tests {
		fs := tc.release.FlagSet()
		assert.Equal(t, tc.expectedName, fs.Name())
	}
}
