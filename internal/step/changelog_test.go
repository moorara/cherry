package step

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChangelogGenerateDry(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		gitHubToken   string
		repo          string
		tag           string
		expectedError string
	}{
		{
			name:        "Success",
			workDir:     os.TempDir(),
			gitHubToken: "github-token",
			repo:        "username/repo",
			tag:         "v0.1.0",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := ChangelogGenerate{
				WorkDir:     tc.workDir,
				GitHubToken: tc.gitHubToken,
				Repo:        tc.repo,
				Tag:         tc.tag,
			}

			ctx := context.Background()
			err := step.Dry(ctx)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, tc.expectedError, err.Error())
			}
		})
	}
}

func TestChangelogGenerateRun(t *testing.T) {
	tests := []struct {
		name             string
		workDir          string
		gitHubToken      string
		repo             string
		tag              string
		expectedError    string
		expectedFilename string
	}{
		{
			name:          "InvalidGitHubToken",
			workDir:       os.TempDir(),
			gitHubToken:   "github-token",
			repo:          "username/repo",
			tag:           "v0.1.0",
			expectedError: `401 - Bad credentials`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := ChangelogGenerate{
				WorkDir:     tc.workDir,
				GitHubToken: tc.gitHubToken,
				Repo:        tc.repo,
				Tag:         tc.tag,
			}

			ctx := context.Background()
			err := step.Run(ctx)

			if tc.expectedError == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedFilename, step.Result.Filename)
				assert.NotEmpty(t, step.Result.Changelog)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			}
		})
	}
}

func TestChangelogGenerateRevert(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		gitHubToken   string
		repo          string
		tag           string
		expectedError string
	}{
		{
			name:          "OK",
			workDir:       os.TempDir(),
			gitHubToken:   "github-token",
			repo:          "username/repo",
			tag:           "v0.1.0",
			expectedError: ``,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := ChangelogGenerate{
				WorkDir:     tc.workDir,
				GitHubToken: tc.gitHubToken,
				Repo:        tc.repo,
				Tag:         tc.tag,
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
