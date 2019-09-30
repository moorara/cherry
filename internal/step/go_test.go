package step

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
			expectedError: "exit status 1: build ./cmd: cannot find module for path ./cmd",
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
			expectedError: "exit status 1: build ./cmd: cannot find module for path ./cmd",
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

			err := step.Run()
			if tc.expectedError == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedPackagePath, step.Result.PackagePath)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
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
				Ctx:        context.Background(),
				LDFlags:    tc.ldflags,
				MainFile:   tc.mainFile,
				BinaryFile: tc.binaryFile,
				Platforms:  tc.platforms,
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
				Ctx:        context.Background(),
				LDFlags:    tc.ldflags,
				MainFile:   tc.mainFile,
				BinaryFile: tc.binaryFile,
				Platforms:  tc.platforms,
			}

			dir, err := ioutil.TempDir("", "cherry-")
			assert.NoError(t, err)
			defer os.RemoveAll(dir)

			step.BinaryFile = filepath.Join(dir, step.BinaryFile)

			err = step.Run()
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
			step := GoBuild{
				Ctx: context.Background(),
			}

			tf, err := ioutil.TempFile("", "cherry-test-")
			assert.NoError(t, err)
			tf.Close()
			defer os.Remove(tf.Name())

			step.Result.Binaries = []string{tf.Name()}

			err = step.Revert()
			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}
