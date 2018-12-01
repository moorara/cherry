package git

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsRepoClean(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		expectedError bool
	}{
		{
			name:          "Success",
			workDir:       ".",
			expectedError: false,
		},
		{
			name:          "Error",
			workDir:       os.TempDir(),
			expectedError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := New(tc.workDir)
			_, err := g.IsRepoClean()

			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetRepoName(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		expectedError bool
	}{
		{
			name:          "Success",
			workDir:       ".",
			expectedError: false,
		},
		{
			name:          "Error",
			workDir:       os.TempDir(),
			expectedError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := New(tc.workDir)
			_, err := g.GetRepoName()

			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
