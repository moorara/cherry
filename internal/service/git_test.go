package service

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGitIsClean(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		expectedError bool
	}{
		{
			name:          "Error",
			workDir:       os.TempDir(),
			expectedError: true,
		},
		{
			name:          "Success",
			workDir:       ".",
			expectedError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			git := NewGit(tc.workDir)
			_, err := git.IsClean()

			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGitGetRepo(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		expectedError bool
	}{
		{
			name:          "Error",
			workDir:       os.TempDir(),
			expectedError: true,
		},
		{
			name:          "Success",
			workDir:       ".",
			expectedError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			git := NewGit(tc.workDir)
			_, err := git.GetRepo()

			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGitGetBranch(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		expectedError bool
	}{
		{
			name:          "Error",
			workDir:       os.TempDir(),
			expectedError: true,
		},
		{
			name:          "Success",
			workDir:       ".",
			expectedError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			git := NewGit(tc.workDir)
			_, err := git.GetBranch()

			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGitGetHEAD(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		expectedError bool
	}{
		{
			name:          "Error",
			workDir:       os.TempDir(),
			expectedError: true,
		},
		{
			name:          "FullSHA",
			workDir:       ".",
			expectedError: false,
		},
		{
			name:          "ShortSHA",
			workDir:       ".",
			expectedError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			git := NewGit(tc.workDir)
			_, err := git.GetHEAD()

			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGitCommit(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		message       string
		files         []string
		expectedError bool
	}{
		{
			name:          "Error",
			workDir:       os.TempDir(),
			message:       "test commit",
			files:         []string{"."},
			expectedError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			git := NewGit(tc.workDir)
			err := git.Commit(tc.message, tc.files...)

			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGitTag(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		tag           string
		expectedError bool
	}{
		{
			name:          "Error",
			workDir:       os.TempDir(),
			tag:           "test-tag",
			expectedError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			git := NewGit(tc.workDir)
			err := git.Tag(tc.tag)

			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGitPush(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		withTags      bool
		expectedError bool
	}{
		{
			name:          "Error",
			workDir:       os.TempDir(),
			withTags:      true,
			expectedError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			git := NewGit(tc.workDir)
			err := git.Push(tc.withTags)

			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRepoPath(t *testing.T) {
	tests := []struct {
		repo         Repo
		expectedPath string
	}{
		{
			repo:         Repo{},
			expectedPath: "/",
		},
		{
			repo: Repo{
				Owner: "moorara",
				Name:  "cherry",
			},
			expectedPath: "moorara/cherry",
		},
	}

	for _, tc := range tests {
		assert.Equal(t, tc.expectedPath, tc.repo.Path())
	}
}
