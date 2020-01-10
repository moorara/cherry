package step

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChangelogGenerateMock(t *testing.T) {
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
			step := ChangelogGenerate{
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

func TestChangelogGenerateDry(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		gitHubToken   string
		gitHubUser    string
		gitHubProject string
		tag           string
		expectedError string
	}{}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := ChangelogGenerate{
				WorkDir:       tc.workDir,
				GitHubToken:   tc.gitHubToken,
				GitHubUser:    tc.gitHubUser,
				GitHubProject: tc.gitHubProject,
				Tag:           tc.tag,
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
		gitHubUser       string
		gitHubProject    string
		tag              string
		expectedError    string
		expectedFilename string
	}{}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := ChangelogGenerate{
				WorkDir:       tc.workDir,
				GitHubToken:   tc.gitHubToken,
				GitHubUser:    tc.gitHubUser,
				GitHubProject: tc.gitHubProject,
				Tag:           tc.tag,
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
		gitHubUser    string
		gitHubProject string
		tag           string
		expectedError string
	}{}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := ChangelogGenerate{
				WorkDir:       tc.workDir,
				GitHubToken:   tc.gitHubToken,
				GitHubUser:    tc.gitHubUser,
				GitHubProject: tc.gitHubProject,
				Tag:           tc.tag,
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
