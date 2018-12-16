package test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTester(t *testing.T) {
	tests := []struct {
		name    string
		workDir string
	}{
		{
			name:    "CurrentPath",
			workDir: "./",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tester := NewTester(tc.workDir)
			assert.NotNil(t, tester)
		})
	}
}

func TestGetPackages(t *testing.T) {
	tests := []struct {
		name             string
		workDir          string
		ctx              context.Context
		expectedError    string
		expectedPackages []string
	}{
		{
			name:             "NoPackage",
			workDir:          os.TempDir(),
			ctx:              context.Background(),
			expectedError:    "exit status 1",
			expectedPackages: nil,
		},
		{
			name:             "CurrentPath",
			workDir:          "./",
			ctx:              context.Background(),
			expectedError:    "",
			expectedPackages: []string{"github.com/moorara/cherry/internal/test"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tester := &tester{
				workDir: tc.workDir,
			}

			packages, err := tester.getPackages(tc.ctx)

			if tc.expectedError != "" {
				assert.Contains(t, err.Error(), tc.expectedError)
				assert.Nil(t, packages)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedPackages, packages)
			}
		})
	}
}

func TestTestPackage(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		ctx           context.Context
		pkg           string
		coverfile     string
		expectedError string
	}{}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tester := &tester{
				workDir: tc.workDir,
			}

			err := tester.testPackage(tc.ctx, tc.pkg, tc.coverfile)

			if tc.expectedError != "" {
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCover(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		ctx           context.Context
		expectedError string
	}{}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tester := &tester{
				workDir: tc.workDir,
			}

			err := tester.Cover(tc.ctx)

			if tc.expectedError != "" {
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				assert.NoError(t, err)
			}

			// Cleanup
			os.RemoveAll(filepath.Join(tc.workDir, reportPath))
		})
	}
}
