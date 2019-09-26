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
				CrossCompile:   defaultCrossCompile,
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
				MainFile:       "cmd/main.go",
				BinaryFile:     "build/app",
				VersionPackage: "./cmd/version",
				GoVersions:     []string{"1.10", "1.11"},
				Platforms:      []string{"linux-amd64", "darwin-amd64", "windows-amd64"},
			},
			Build{
				CrossCompile:   true,
				MainFile:       "cmd/main.go",
				BinaryFile:     "build/app",
				VersionPackage: "./cmd/version",
				GoVersions:     []string{"1.10", "1.11"},
				Platforms:      []string{"linux-amd64", "darwin-amd64", "windows-amd64"},
			},
		},
	}

	for _, tc := range tests {
		tc.build.SetDefaults()
		assert.Equal(t, tc.expectedBuild, tc.build)
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
				VersionPackage: "./cmd/version",
				GoVersions:     []string{"1.10", "1.11"},
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

func TestReleaseSetDefaults(t *testing.T) {
	tests := []struct {
		release         Release
		expectedRelease Release
	}{
		{
			Release{},
			Release{
				Model: defaultModel,
				Build: defaultBuild,
			},
		},
		{
			Release{
				Model: "branch",
				Build: true,
			},
			Release{
				Model: "branch",
				Build: true,
			},
		},
	}

	for _, tc := range tests {
		tc.release.SetDefaults()
		assert.Equal(t, tc.expectedRelease, tc.release)
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
				Model: "master",
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

func TestTestSetDefaults(t *testing.T) {
	tests := []struct {
		test         Test
		expectedTest Test
	}{
		{
			Test{},
			Test{
				CoverMode:  defaultCoverMode,
				ReportPath: defaultReportPath,
			},
		},
		{
			Test{
				CoverMode:  "set",
				ReportPath: "report",
			},
			Test{
				CoverMode:  "set",
				ReportPath: "report",
			},
		},
	}

	for _, tc := range tests {
		tc.test.SetDefaults()
		assert.Equal(t, tc.expectedTest, tc.test)
	}
}

func TestTestFlagSet(t *testing.T) {
	tests := []struct {
		test         Test
		expectedName string
	}{
		{
			test:         Test{},
			expectedName: "test",
		},
		{
			test: Test{
				CoverMode:  "atomic",
				ReportPath: "coverage",
			},
			expectedName: "test",
		},
	}

	for _, tc := range tests {
		fs := tc.test.FlagSet()
		assert.Equal(t, tc.expectedName, fs.Name())
	}
}

func TestSpecSetDefaults(t *testing.T) {
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
				VersionFile: defaultVersionFile,
				Test: Test{
					CoverMode:  defaultCoverMode,
					ReportPath: defaultReportPath,
				},
				Build: Build{
					CrossCompile:   defaultCrossCompile,
					MainFile:       defaultMainFile,
					BinaryFile:     "bin/spec",
					VersionPackage: defaultVersionPackage,
					GoVersions:     defaultGoVersions,
					Platforms:      defaultPlatforms,
				},
				Release: Release{
					Model: defaultModel,
					Build: defaultBuild,
				},
			},
		},
		{
			Spec{
				Version:     "2.0",
				Language:    "go",
				VersionFile: "version.yaml",
				Test: Test{
					CoverMode:  "atomic",
					ReportPath: "report",
				},
				Build: Build{
					CrossCompile:   true,
					MainFile:       "cmd/main.go",
					BinaryFile:     "build/app",
					VersionPackage: "./cmd/version",
					GoVersions:     []string{"1.10", "1.11"},
					Platforms:      []string{"linux-amd64", "darwin-amd64", "windows-amd64"},
				},
				Release: Release{
					Model: "branch",
					Build: true,
				},
			},
			Spec{
				ToolName:    defaultToolName,
				ToolVersion: "",
				Version:     "2.0",
				Language:    "go",
				VersionFile: "version.yaml",
				Test: Test{
					CoverMode:  "atomic",
					ReportPath: "report",
				},
				Build: Build{
					CrossCompile:   true,
					MainFile:       "cmd/main.go",
					BinaryFile:     "build/app",
					VersionPackage: "./cmd/version",
					GoVersions:     []string{"1.10", "1.11"},
					Platforms:      []string{"linux-amd64", "darwin-amd64", "windows-amd64"},
				},
				Release: Release{
					Model: "branch",
					Build: true,
				},
			},
		},
	}

	for _, tc := range tests {
		tc.spec.SetDefaults()
		assert.Equal(t, tc.expectedSpec, tc.spec)
	}
}

func TestReadSpec(t *testing.T) {
	tests := []struct {
		name          string
		specFile      string
		expectedSpec  *Spec
		expectedError string
	}{
		{
			name:          "BadFile",
			specFile:      "test/null",
			expectedError: "no such file or directory",
		},
		{
			name:          "UnknownFile",
			specFile:      "test/unknown.hcl",
			expectedError: "unknown spec file",
		},
		{
			name:          "EmptyYAML",
			specFile:      "test/empty.yaml",
			expectedError: "EOF",
		},
		{
			name:          "EmptyJSON",
			specFile:      "test/empty.json",
			expectedError: "EOF",
		},
		{
			name:          "InvalidYAML",
			specFile:      "test/invalid.yaml",
			expectedError: "cannot unmarshal",
		},
		{
			name:          "InvalidJSON",
			specFile:      "test/invalid.json",
			expectedError: "invalid character",
		},
		{
			name:     "MinimumYAML",
			specFile: "test/min.yaml",
			expectedSpec: &Spec{
				Version:  "1.0",
				Language: "go",
				Test:     Test{},
				Build:    Build{},
				Release: Release{
					Build: true,
				},
			},
		},
		{
			name:     "MinimumJSON",
			specFile: "test/min.json",
			expectedSpec: &Spec{
				Version:  "1.0",
				Language: "go",
				Test:     Test{},
				Build:    Build{},
				Release: Release{
					Build: true,
				},
			},
		},
		{
			name:     "MaximumYAML",
			specFile: "test/max.yaml",
			expectedSpec: &Spec{
				Version:     "1.0",
				Language:    "go",
				VersionFile: "VERSION",
				Test: Test{
					CoverMode:  "atomic",
					ReportPath: "coverage",
				},
				Build: Build{
					CrossCompile:   true,
					MainFile:       "main.go",
					BinaryFile:     "bin/cherry",
					VersionPackage: "./cmd/version",
					GoVersions:     []string{"1.11", "1.12.10", "1.13.1"},
					Platforms:      []string{"linux-386", "linux-amd64", "linux-arm", "linux-arm64", "darwin-386", "darwin-amd64", "windows-386", "windows-amd64"},
				},
				Release: Release{
					Model: "master",
					Build: true,
				},
			},
		},
		{
			name:     "MaximumJSON",
			specFile: "test/max.json",
			expectedSpec: &Spec{
				Version:     "1.0",
				Language:    "go",
				VersionFile: "VERSION",
				Test: Test{
					CoverMode:  "atomic",
					ReportPath: "coverage",
				},
				Build: Build{
					CrossCompile:   true,
					MainFile:       "main.go",
					BinaryFile:     "bin/cherry",
					VersionPackage: "./cmd/version",
					GoVersions:     []string{"1.11", "1.12.10", "1.13.1"},
					Platforms:      []string{"linux-386", "linux-amd64", "linux-arm", "linux-arm64", "darwin-386", "darwin-amd64", "windows-386", "windows-amd64"},
				},
				Release: Release{
					Model: "master",
					Build: true,
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			specFiles = []string{tc.specFile}
			spec, err := ReadSpec()

			if tc.expectedError != "" {
				assert.Contains(t, err.Error(), tc.expectedError)
				assert.Nil(t, spec)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedSpec, spec)
			}
		})
	}

	t.Run("NoFile", func(t *testing.T) {
		specFiles = []string{}
		spec, err := ReadSpec()

		assert.Equal(t, "no spec file found", err.Error())
		assert.Nil(t, spec)
	})
}
