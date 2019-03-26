package service

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseGitRepo(t *testing.T) {
	tests := []struct {
		name          string
		output        string
		expectedOwner string
		expectedName  string
		expectedError string
	}{
		{
			name:          "Empty",
			output:        ``,
			expectedOwner: "",
			expectedName:  "",
			expectedError: "failed to get git repository url",
		},
		{
			name: "Invalid",
			output: `
			origin	moorara/cherry (fetch)
			origin	moorara/cherry (push)
			`,
			expectedOwner: "",
			expectedName:  "",
			expectedError: "failed to get git repository name",
		},
		{
			name: "HTTPSSchema",
			output: `
			origin	https://github.com/moorara/cherry (fetch)
			origin	https://github.com/moorara/cherry (push)
			`,
			expectedOwner: "moorara",
			expectedName:  "cherry",
			expectedError: "",
		},
		{
			name: "HTTPSSchemaWithGit",
			output: `
			origin	https://github.com/moorara/cherry.git (fetch)
			origin	https://github.com/moorara/cherry.git (push)
			`,
			expectedOwner: "moorara",
			expectedName:  "cherry",
			expectedError: "",
		},
		{
			name: "SSHSchema",
			output: `
			origin	git@github.com:moorara/cherry (fetch)
			origin	git@github.com:moorara/cherry (push)
			`,
			expectedOwner: "moorara",
			expectedName:  "cherry",
			expectedError: "",
		},
		{
			name: "SSHSchemaWithGit",
			output: `
			origin	git@github.com:moorara/cherry.git (fetch)
			origin	git@github.com:moorara/cherry.git (push)
			`,
			expectedOwner: "moorara",
			expectedName:  "cherry",
			expectedError: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			owner, name, err := parseGitRepo(tc.output)

			if tc.expectedError != "" {
				assert.Equal(t, tc.expectedError, err.Error())
				assert.Empty(t, owner)
				assert.Empty(t, name)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedOwner, owner)
				assert.Equal(t, tc.expectedName, name)
			}
		})
	}
}

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
		annotation    *Annotation
		expectedError bool
	}{
		{
			name:          "Error",
			workDir:       os.TempDir(),
			tag:           "test-tag",
			annotation:    nil,
			expectedError: true,
		},
		{
			name:          "Error",
			workDir:       os.TempDir(),
			tag:           "test-tag",
			annotation:    &Annotation{Message: "tag message"},
			expectedError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			git := NewGit(tc.workDir)
			err := git.Tag(tc.tag, tc.annotation)

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
		expectedError bool
	}{
		{
			name:          "Error",
			workDir:       os.TempDir(),
			expectedError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			git := NewGit(tc.workDir)
			err := git.Push()

			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGitPushTag(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		tag           string
		expectedError bool
	}{
		{
			name:          "Error",
			workDir:       os.TempDir(),
			tag:           "v0.1.0",
			expectedError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			git := NewGit(tc.workDir)
			err := git.PushTag(tc.tag)

			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGitPull(t *testing.T) {
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
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			git := NewGit(tc.workDir)
			err := git.Pull()

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
