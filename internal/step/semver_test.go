package step

import (
	"context"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/moorara/cherry/pkg/semver"
	"github.com/stretchr/testify/assert"
)

func TestSemVerReadMock(t *testing.T) {
	tests := []struct {
		name                string
		mock                *mockStep
		expectedDryError    error
		expectedRunError    error
		expectedRevertError error
	}{
		{
			name: "OK",
			mock: &mockStep{},
		},
		{
			name: "OK",
			mock: &mockStep{
				DryOutError:    errors.New("dry error"),
				RunOutError:    errors.New("run error"),
				RevertOutError: errors.New("revert error"),
			},
			expectedDryError:    errors.New("dry error"),
			expectedRunError:    errors.New("run error"),
			expectedRevertError: errors.New("revert error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := SemVerRead{
				Mock: tc.mock,
			}

			ctx := context.Background()

			err := step.Dry(ctx)
			assert.Equal(t, tc.expectedDryError, err)

			err = step.Run(ctx)
			assert.Equal(t, tc.expectedRunError, err)

			err = step.Revert(ctx)
			assert.Equal(t, tc.expectedRevertError, err)
		})
	}
}

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
			expectedError: `SemVerRead.Dry: no version file`,
		},
		{
			name:          "MissingVersionFile",
			workDir:       "./test",
			filename:      "missing",
			expectedError: `SemVerRead.Dry: open test/missing: no such file or directory`,
		},
		{
			name:          "EmptyTextFile",
			workDir:       "./test",
			filename:      "empty",
			expectedError: `SemVerRead.Dry: empty version`,
		},
		{
			name:          "EmptyJSONFile",
			workDir:       "./test",
			filename:      "empty.json",
			expectedError: `SemVerRead.Dry: unexpected end of JSON input`,
		},
		{
			name:          "NoJSONVersion",
			workDir:       "./test",
			filename:      "noversion.json",
			expectedError: `SemVerRead.Dry: empty version`,
		},
		{
			name:          "InvalidTextVersion",
			workDir:       "./test",
			filename:      "invalid",
			expectedError: `SemVerRead.Dry: invalid semantic version`,
		},
		{
			name:          "InvalidJSONVersion",
			workDir:       "./test",
			filename:      "invalid.json",
			expectedError: `SemVerRead.Dry: invalid semantic version`,
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

			ctx := context.Background()
			err := step.Dry(ctx)

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
		name             string
		workDir          string
		filename         string
		expectedError    string
		expectedFilename string
		expectedSemver   semver.SemVer
	}{
		{
			name:          "NoVersionFile",
			workDir:       "./",
			filename:      "",
			expectedError: `SemVerRead.Run: no version file`,
		},
		{
			name:          "MissingVersionFile",
			workDir:       "./test",
			filename:      "missing",
			expectedError: `SemVerRead.Run: open test/missing: no such file or directory`,
		},
		{
			name:          "EmptyTextFile",
			workDir:       "./test",
			filename:      "empty",
			expectedError: `SemVerRead.Run: empty version`,
		},
		{
			name:          "EmptyJSONFile",
			workDir:       "./test",
			filename:      "empty.json",
			expectedError: `SemVerRead.Run: unexpected end of JSON input`,
		},
		{
			name:          "NoJSONVersion",
			workDir:       "./test",
			filename:      "noversion.json",
			expectedError: `SemVerRead.Run: empty version`,
		},
		{
			name:          "InvalidTextVersion",
			workDir:       "./test",
			filename:      "invalid",
			expectedError: `SemVerRead.Run: invalid semantic version`,
		},
		{
			name:          "InvalidJSONVersion",
			workDir:       "./test",
			filename:      "invalid.json",
			expectedError: `SemVerRead.Run: invalid semantic version`,
		},
		{
			name:             "TextFileSuccess",
			workDir:          "./test",
			filename:         "VERSION",
			expectedFilename: "VERSION",
			expectedSemver: semver.SemVer{
				Major: 0,
				Minor: 1,
				Patch: 0,
			},
		},
		{
			name:             "JSONFileSuccess",
			workDir:          "./test",
			filename:         "package.json",
			expectedFilename: "package.json",
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

			ctx := context.Background()
			err := step.Run(ctx)

			if tc.expectedError == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedFilename, step.Result.Filename)
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

			ctx := context.Background()
			err := step.Revert(ctx)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}

func TestSemVerUpdateMock(t *testing.T) {
	tests := []struct {
		name                string
		mock                *mockStep
		expectedDryError    error
		expectedRunError    error
		expectedRevertError error
	}{
		{
			name: "OK",
			mock: &mockStep{},
		},
		{
			name: "OK",
			mock: &mockStep{
				DryOutError:    errors.New("dry error"),
				RunOutError:    errors.New("run error"),
				RevertOutError: errors.New("revert error"),
			},
			expectedDryError:    errors.New("dry error"),
			expectedRunError:    errors.New("run error"),
			expectedRevertError: errors.New("revert error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := SemVerUpdate{
				Mock: tc.mock,
			}

			ctx := context.Background()

			err := step.Dry(ctx)
			assert.Equal(t, tc.expectedDryError, err)

			err = step.Run(ctx)
			assert.Equal(t, tc.expectedRunError, err)

			err = step.Revert(ctx)
			assert.Equal(t, tc.expectedRevertError, err)
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
			expectedError: `SemVerUpdate.Dry: no version file`,
		},
		{
			name:          "MissingVersionFile",
			workDir:       "./test",
			filename:      "missing",
			version:       "0.2.0",
			expectedError: `SemVerUpdate.Dry: version file not found`,
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

			ctx := context.Background()
			err := step.Dry(ctx)

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
		name             string
		mockFilename     string
		version          string
		expectedFilename string
		expectedError    string
	}{
		{
			name:          "NoVersionFile",
			version:       "0.2.0",
			expectedError: `SemVerUpdate.Run: no version file`,
		},
		{
			name:             "TextFileSuccess",
			mockFilename:     "VERSION",
			version:          "0.2.0",
			expectedFilename: "VERSION",
		},
		{
			name:             "JSONFileSuccess",
			mockFilename:     "package.json",
			version:          "0.2.0",
			expectedFilename: "package.json",
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

			ctx := context.Background()
			err := step.Run(ctx)

			if tc.expectedError == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedFilename, step.Result.Filename)
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

			ctx := context.Background()
			err := step.Revert(ctx)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}
