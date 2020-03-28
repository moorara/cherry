package step

import (
	"context"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGoVersionMock(t *testing.T) {
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
			step := GoVersion{
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

func TestGoVersionDry(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		expectedError string
	}{
		{
			name:    "Success",
			workDir: "./test",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GoVersion{
				WorkDir: tc.workDir,
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

func TestGoVersionRun(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		expectedError string
	}{
		{
			name:    "Success",
			workDir: "./test",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GoVersion{
				WorkDir: tc.workDir,
			}

			ctx := context.Background()
			err := step.Run(ctx)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}

func TestGoVersionRevert(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		expectedError string
	}{
		{
			name:    "Success",
			workDir: "./test",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GoVersion{
				WorkDir: tc.workDir,
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

func TestGoListMock(t *testing.T) {
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
			step := GoList{
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

func TestGoListDry(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		pkg           string
		expectedError string
	}{
		{
			name:          "InvalidPackage",
			workDir:       "./test",
			pkg:           "./cmd",
			expectedError: "GoList.Dry: exit status 1 can't load package: package ./cmd",
		},
		{
			name:    "Success",
			workDir: "./test",
			pkg:     ".",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GoList{
				WorkDir: tc.workDir,
				Package: tc.pkg,
			}

			ctx := context.Background()
			err := step.Dry(ctx)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			}
		})
	}
}

func TestGoListRun(t *testing.T) {
	tests := []struct {
		name                string
		workDir             string
		pkg                 string
		expectedError       string
		expectedPackagePath string
	}{
		{
			name:          "InvalidPackage",
			workDir:       "./test",
			pkg:           "./cmd",
			expectedError: "GoList.Run: exit status 1 can't load package: package ./cmd",
		},
		{
			name:                "Success",
			workDir:             "./test",
			pkg:                 ".",
			expectedPackagePath: "github.com/moorara/cherry/internal/step/test",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GoList{
				WorkDir: tc.workDir,
				Package: tc.pkg,
			}

			ctx := context.Background()
			err := step.Run(ctx)

			if tc.expectedError == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedPackagePath, step.Result.PackagePath)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			}
		})
	}
}

func TestGoListRevert(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		pkg           string
		expectedError string
	}{
		{
			name:    "Success",
			workDir: "./test",
			pkg:     ".",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GoList{
				WorkDir: tc.workDir,
				Package: tc.pkg,
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

func TestGoBuildMock(t *testing.T) {
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
			step := GoBuild{
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

func TestGoBuildDry(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		ldflags       string
		mainFile      string
		binaryFile    string
		platforms     []string
		expectedError string
	}{
		{
			name:       "Success",
			workDir:    "./test",
			ldflags:    "",
			mainFile:   "main.go",
			binaryFile: "app",
			platforms:  nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GoBuild{
				WorkDir:    tc.workDir,
				LDFlags:    tc.ldflags,
				MainFile:   tc.mainFile,
				BinaryFile: tc.binaryFile,
				Platforms:  tc.platforms,
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

func TestGoBuildRun(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		ldflags       string
		mainFile      string
		binaryFile    string
		platforms     []string
		expectedError string
	}{
		{
			name:       "Success",
			workDir:    "./test",
			ldflags:    "",
			mainFile:   "main.go",
			binaryFile: "app",
			platforms:  nil,
		},
		{
			name:       "CrossCompile",
			workDir:    "./test",
			ldflags:    "",
			mainFile:   "main.go",
			binaryFile: "app",
			platforms:  []string{"linux-amd64"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GoBuild{
				WorkDir:    tc.workDir,
				LDFlags:    tc.ldflags,
				MainFile:   tc.mainFile,
				BinaryFile: tc.binaryFile,
				Platforms:  tc.platforms,
			}

			dir, err := ioutil.TempDir("", "cherry-")
			assert.NoError(t, err)
			defer os.RemoveAll(dir)

			step.BinaryFile = filepath.Join(dir, step.BinaryFile)

			ctx := context.Background()
			err = step.Run(ctx)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}

func TestGoBuildRevert(t *testing.T) {
	tests := []struct {
		name          string
		expectedError string
	}{
		{
			name: "Success",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GoBuild{}

			tf, err := ioutil.TempFile("", "cherry-test-")
			assert.NoError(t, err)
			tf.Close()
			defer os.Remove(tf.Name())

			step.Result.Binaries = []string{tf.Name()}

			ctx := context.Background()
			err = step.Revert(ctx)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}
