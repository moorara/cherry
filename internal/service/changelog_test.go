package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewChangelog(t *testing.T) {
	tests := []struct {
		workDir     string
		githubToken string
	}{
		{
			workDir:     ".",
			githubToken: "github-token",
		},
	}

	for _, tc := range tests {
		changelog := NewChangelog(tc.workDir, tc.githubToken)
		assert.NotNil(t, changelog)
	}
}

func TestChangelogFilename(t *testing.T) {
	tests := []struct {
		workDir          string
		githubToken      string
		expectedFilename string
	}{
		{
			workDir:          ".",
			githubToken:      "github-token",
			expectedFilename: changelogFilename,
		},
	}

	for _, tc := range tests {
		changelog := &changelog{
			workDir:     tc.workDir,
			githubToken: tc.githubToken,
		}

		filename := changelog.Filename()
		assert.Equal(t, tc.expectedFilename, filename)
	}
}

func TestChangelogGenerate(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		githubToken   string
		ctx           context.Context
		repo          string
		tag           string
		expectedError string
	}{}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			changelog := &changelog{
				workDir:     tc.workDir,
				githubToken: tc.githubToken,
			}

			_, err := changelog.Generate(tc.ctx, tc.repo, tc.tag)

			if tc.expectedError != "" {
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
