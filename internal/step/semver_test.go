package step

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/moorara/cherry/internal/semver"
	"github.com/stretchr/testify/assert"
)

func TestSemVerReadDry(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		filename      string
		expectedError string
	}{
		{
			name:          "NoVersionFile",
			workDir:       "./",
			filename:      "",
			expectedError: `no version file`,
		},
		{
			name:          "MissingVersionFile",
			workDir:       "./test",
			filename:      "missing",
			expectedError: `open test/missing: no such file or directory`,
		},
		{
			name:          "EmptyTextFile",
			workDir:       "./test",
			filename:      "empty",
			expectedError: `empty version`,
		},
		{
			name:          "EmptyJSONFile",
			workDir:       "./test",
			filename:      "empty.json",
			expectedError: `unexpected end of JSON input`,
		},
		{
			name:          "NoJSONVersion",
			workDir:       "./test",
			filename:      "noversion.json",
			expectedError: `empty version`,
		},
		{
			name:          "InvalidTextVersion",
			workDir:       "./test",
			filename:      "invalid",
			expectedError: `invalid major version: strconv.ParseUint: parsing "x": invalid syntax`,
		},
		{
			name:          "InvalidJSONVersion",
			workDir:       "./test",
			filename:      "invalid.json",
			expectedError: `invalid major version: strconv.ParseUint: parsing "x": invalid syntax`,
		},
		{
			name:     "TextFileSuccess",
			workDir:  "./test",
			filename: "VERSION",
		},
		{
			name:     "JSONFileSuccess",
			workDir:  "./test",
			filename: "package.json",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := SemVerRead{
				WorkDir:  tc.workDir,
				Filename: tc.filename,
			}

			err := step.Dry()
			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}

func TestSemVerReadRun(t *testing.T) {
	tests := []struct {
		name           string
		workDir        string
		filename       string
		expectedError  string
		expectedSemver semver.SemVer
	}{
		{
			name:          "NoVersionFile",
			workDir:       "./",
			filename:      "",
			expectedError: `no version file`,
		},
		{
			name:          "MissingVersionFile",
			workDir:       "./test",
			filename:      "missing",
			expectedError: `open test/missing: no such file or directory`,
		},
		{
			name:          "EmptyTextFile",
			workDir:       "./test",
			filename:      "empty",
			expectedError: `empty version`,
		},
		{
			name:          "EmptyJSONFile",
			workDir:       "./test",
			filename:      "empty.json",
			expectedError: `unexpected end of JSON input`,
		},
		{
			name:          "NoJSONVersion",
			workDir:       "./test",
			filename:      "noversion.json",
			expectedError: `empty version`,
		},
		{
			name:          "InvalidTextVersion",
			workDir:       "./test",
			filename:      "invalid",
			expectedError: `invalid major version: strconv.ParseUint: parsing "x": invalid syntax`,
		},
		{
			name:          "InvalidJSONVersion",
			workDir:       "./test",
			filename:      "invalid.json",
			expectedError: `invalid major version: strconv.ParseUint: parsing "x": invalid syntax`,
		},
		{
			name:     "TextFileSuccess",
			workDir:  "./test",
			filename: "VERSION",
			expectedSemver: semver.SemVer{
				Major: 0,
				Minor: 1,
				Patch: 0,
			},
		},
		{
			name:     "JSONFileSuccess",
			workDir:  "./test",
			filename: "package.json",
			expectedSemver: semver.SemVer{
				Major: 0,
				Minor: 1,
				Patch: 0,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := SemVerRead{
				WorkDir:  tc.workDir,
				Filename: tc.filename,
			}

			err := step.Run()
			if tc.expectedError == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedSemver, step.Result.Version)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}

func TestSemVerReadRevert(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		filename      string
		expectedError string
	}{
		{
			name:     "TextFileSuccess",
			workDir:  "./test",
			filename: "VERSION",
		},
		{
			name:     "JSONFileSuccess",
			workDir:  "./test",
			filename: "package.json",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := SemVerRead{
				WorkDir:  tc.workDir,
				Filename: tc.filename,
			}

			err := step.Revert()
			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}

func TestSemVerUpdateDry(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		filename      string
		version       string
		expectedError string
	}{
		{
			name:          "NoVersionFile",
			workDir:       "./",
			filename:      "",
			version:       "0.2.0",
			expectedError: `no version file`,
		},
		{
			name:          "MissingVersionFile",
			workDir:       "./test",
			filename:      "missing",
			version:       "0.2.0",
			expectedError: `version file not found`,
		},
		{
			name:     "TextFileSuccess",
			workDir:  "./test",
			filename: "VERSION",
			version:  "0.2.0",
		},
		{
			name:     "JSONFileSuccess",
			workDir:  "./test",
			filename: "package.json",
			version:  "0.2.0",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := SemVerUpdate{
				WorkDir:  tc.workDir,
				Filename: tc.filename,
				Version:  tc.version,
			}

			err := step.Dry()
			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}

func TestSemVerUpdateRun(t *testing.T) {
	tests := []struct {
		name          string
		mockFilename  string
		version       string
		expectedError string
	}{
		{
			name:          "NoVersionFile",
			version:       "0.2.0",
			expectedError: `no version file`,
		},
		{
			name:         "TextFileSuccess",
			mockFilename: "VERSION",
			version:      "0.2.0",
		},
		{
			name:         "JSONFileSuccess",
			mockFilename: "package.json",
			version:      "0.2.0",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := SemVerUpdate{
				Version: tc.version,
			}

			if tc.mockFilename != "" {
				td, err := ioutil.TempDir("", "cherry-")
				assert.NoError(t, err)
				defer os.RemoveAll(td)

				tf := filepath.Join(td, tc.mockFilename)
				err = ioutil.WriteFile(tf, []byte(""), 0644)
				assert.NoError(t, err)
				defer os.Remove(tf)

				step.WorkDir = td
				step.Filename = tc.mockFilename
			}

			err := step.Run()
			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}

func TestSemVerUpdateRevert(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		filename      string
		version       string
		expectedError string
	}{
		{
			name:     "TextFileSuccess",
			workDir:  "./test",
			filename: "VERSION",
			version:  "0.2.0",
		},
		{
			name:     "JSONFileSuccess",
			workDir:  "./test",
			filename: "package.json",
			version:  "0.2.0",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := SemVerUpdate{
				WorkDir:  tc.workDir,
				Filename: tc.filename,
				Version:  tc.version,
			}

			err := step.Revert()
			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}