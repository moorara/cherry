package step

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseGitURL(t *testing.T) {
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
			owner, name, err := parseGitURL(tc.output)

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

func TestGitStatus(t *testing.T) {
	tests := []struct {
		name     string
		workDir  string
		dryError string
		runError string
	}{
		{
			name:     "Error",
			workDir:  os.TempDir(),
			dryError: `exit status 128: fatal: not a git repository (or any of the parent directories): .git`,
			runError: `exit status 128: fatal: not a git repository (or any of the parent directories): .git`,
		},
		{
			name:     "Success",
			workDir:  ".",
			dryError: "",
			runError: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitStatus{
				WorkDir: tc.workDir,
			}

			// Test Dry
			err := step.Dry()
			if tc.dryError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.dryError, err.Error())
			}

			// Test Run
			err = step.Run()
			if tc.runError == "" {
				assert.NoError(t, err)
				// TODO: Test results (IsClean)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.runError, err.Error())
			}

			// Test Revert
			err = step.Revert()
			assert.NoError(t, err)
		})
	}
}

func TestGitGetRepo(t *testing.T) {
	tests := []struct {
		name     string
		workDir  string
		dryError string
		runError string
	}{
		{
			name:     "Error",
			workDir:  os.TempDir(),
			dryError: `exit status 128: fatal: not a git repository (or any of the parent directories): .git`,
			runError: `exit status 128: fatal: not a git repository (or any of the parent directories): .git`,
		},
		{
			name:     "Success",
			workDir:  ".",
			dryError: "",
			runError: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitGetRepo{
				WorkDir: tc.workDir,
			}

			// Test Dry
			err := step.Dry()
			if tc.dryError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.dryError, err.Error())
			}

			// Test Run
			err = step.Run()
			if tc.runError == "" {
				assert.NoError(t, err)
				assert.NotEmpty(t, step.Result.Owner)
				assert.NotEmpty(t, step.Result.Name)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.runError, err.Error())
			}

			// Test Revert
			err = step.Revert()
			assert.NoError(t, err)
		})
	}
}

func TestGitGetBranch(t *testing.T) {
	tests := []struct {
		name     string
		workDir  string
		dryError string
		runError string
	}{
		{
			name:     "Error",
			workDir:  os.TempDir(),
			dryError: `exit status 128: fatal: not a git repository (or any of the parent directories): .git`,
			runError: `exit status 128: fatal: not a git repository (or any of the parent directories): .git`,
		},
		{
			name:     "Success",
			workDir:  ".",
			dryError: "",
			runError: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitGetBranch{
				WorkDir: tc.workDir,
			}

			// Test Dry
			err := step.Dry()
			if tc.dryError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.dryError, err.Error())
			}

			// Test Run
			err = step.Run()
			if tc.runError == "" {
				assert.NoError(t, err)
				assert.NotEmpty(t, step.Result.Name)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.runError, err.Error())
			}

			// Test Revert
			err = step.Revert()
			assert.NoError(t, err)
		})
	}
}

func TestGitGetHEAD(t *testing.T) {
	tests := []struct {
		name     string
		workDir  string
		dryError string
		runError string
	}{
		{
			name:     "Error",
			workDir:  os.TempDir(),
			dryError: `exit status 128: fatal: not a git repository (or any of the parent directories): .git`,
			runError: `exit status 128: fatal: not a git repository (or any of the parent directories): .git`,
		},
		{
			name:     "FullSHA",
			workDir:  ".",
			dryError: "",
			runError: "",
		},
		{
			name:     "ShortSHA",
			workDir:  ".",
			dryError: "",
			runError: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitGetHEAD{
				WorkDir: tc.workDir,
			}

			// Test Dry
			err := step.Dry()
			if tc.dryError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.dryError, err.Error())
			}

			// Test Run
			err = step.Run()
			if tc.runError == "" {
				assert.NoError(t, err)
				assert.Len(t, step.Result.SHA, 40)
				assert.Len(t, step.Result.ShortSHA, 7)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.runError, err.Error())
			}

			// Test Revert
			err = step.Revert()
			assert.NoError(t, err)
		})
	}
}

func TestGitAdd(t *testing.T) {
	tests := []struct {
		name        string
		workDir     string
		files       []string
		dryError    string
		runError    string
		revertError string
	}{
		{
			name:        "Error",
			workDir:     os.TempDir(),
			files:       []string{"."},
			dryError:    `exit status 128: fatal: not a git repository (or any of the parent directories): .git`,
			runError:    `exit status 128: fatal: not a git repository (or any of the parent directories): .git`,
			revertError: `exit status 128: fatal: not a git repository (or any of the parent directories): .git`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitAdd{
				WorkDir: tc.workDir,
				Files:   tc.files,
			}

			// Test Dry
			err := step.Dry()
			if tc.dryError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.dryError, err.Error())
			}

			// Test Run
			err = step.Run()
			if tc.runError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.runError, err.Error())
			}

			// Test Revert
			err = step.Revert()
			if tc.revertError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.revertError, err.Error())
			}
		})
	}
}

func TestGitCommit(t *testing.T) {
	tests := []struct {
		name        string
		workDir     string
		message     string
		dryError    string
		runError    string
		revertError string
	}{
		{
			name:        "Error",
			workDir:     os.TempDir(),
			message:     "test message",
			dryError:    `exit status 128: fatal: not a git repository (or any of the parent directories): .git`,
			runError:    `exit status 128: fatal: not a git repository (or any of the parent directories): .git`,
			revertError: `exit status 128: fatal: not a git repository (or any of the parent directories): .git`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitCommit{
				WorkDir: tc.workDir,
				Message: tc.message,
			}

			// Test Dry
			err := step.Dry()
			if tc.dryError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.dryError, err.Error())
			}

			// Test Run
			err = step.Run()
			if tc.runError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.runError, err.Error())
			}

			// Test Revert
			err = step.Revert()
			if tc.revertError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.revertError, err.Error())
			}
		})
	}
}

func TestGitTag(t *testing.T) {
	tests := []struct {
		name        string
		workDir     string
		tag         string
		annotation  string
		dryError    string
		runError    string
		revertError string
	}{
		{
			name:        "Error",
			workDir:     os.TempDir(),
			tag:         "test-tag",
			annotation:  "",
			dryError:    "exit status 128: fatal: not a git repository (or any of the parent directories): .git",
			runError:    "exit status 128: fatal: not a git repository (or any of the parent directories): .git",
			revertError: "exit status 128: fatal: not a git repository (or any of the parent directories): .git",
		},
		{
			name:        "Error",
			workDir:     os.TempDir(),
			tag:         "test-tag",
			annotation:  "annotation message",
			dryError:    `exit status 128: fatal: not a git repository (or any of the parent directories): .git`,
			runError:    `exit status 128: fatal: not a git repository (or any of the parent directories): .git`,
			revertError: `exit status 128: fatal: not a git repository (or any of the parent directories): .git`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitTag{
				WorkDir:    tc.workDir,
				Tag:        tc.tag,
				Annotation: tc.annotation,
			}

			// Test Dry
			err := step.Dry()
			if tc.dryError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.dryError, err.Error())
			}

			// Test Run
			err = step.Run()
			if tc.runError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.runError, err.Error())
			}

			// Test Revert
			err = step.Revert()
			if tc.revertError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.revertError, err.Error())
			}
		})
	}
}

func TestGitPush(t *testing.T) {
	tests := []struct {
		name        string
		workDir     string
		dryError    string
		runError    string
		revertError string
	}{
		{
			name:        "Error",
			workDir:     os.TempDir(),
			dryError:    `exit status 128: fatal: not a git repository (or any of the parent directories): .git`,
			runError:    `exit status 128: fatal: not a git repository (or any of the parent directories): .git`,
			revertError: `cannot revert git push`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitPush{
				WorkDir: tc.workDir,
			}

			// Test Dry
			err := step.Dry()
			if tc.dryError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.dryError, err.Error())
			}

			// Test Run
			err = step.Run()
			if tc.runError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.runError, err.Error())
			}

			// Test Revert
			err = step.Revert()
			if tc.revertError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.revertError, err.Error())
			}
		})
	}
}

func TestGitPushTag(t *testing.T) {
	tests := []struct {
		name        string
		workDir     string
		tag         string
		dryError    string
		runError    string
		revertError string
	}{
		{
			name:        "Error",
			workDir:     os.TempDir(),
			tag:         "v0.1.0",
			dryError:    `exit status 128: fatal: not a git repository (or any of the parent directories): .git`,
			runError:    `exit status 128: fatal: not a git repository (or any of the parent directories): .git`,
			revertError: `cannot revert git push`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitPushTag{
				WorkDir: tc.workDir,
				Tag:     tc.tag,
			}

			// Test Dry
			err := step.Dry()
			if tc.dryError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.dryError, err.Error())
			}

			// Test Run
			err = step.Run()
			if tc.runError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.runError, err.Error())
			}

			// Test Revert
			err = step.Revert()
			if tc.revertError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.revertError, err.Error())
			}
		})
	}
}

func TestGitPull(t *testing.T) {
	tests := []struct {
		name        string
		workDir     string
		dryError    string
		runError    string
		revertError string
	}{
		{
			name:        "Error",
			workDir:     os.TempDir(),
			dryError:    `exit status 128: fatal: not a git repository (or any of the parent directories): .git`,
			runError:    `exit status 128: fatal: not a git repository (or any of the parent directories): .git`,
			revertError: `cannot revert git pull`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitPull{
				WorkDir: tc.workDir,
			}

			// Test Dry
			err := step.Dry()
			if tc.dryError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.dryError, err.Error())
			}

			// Test Run
			err = step.Run()
			if tc.runError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.runError, err.Error())
			}

			// Test Revert
			err = step.Revert()
			if tc.revertError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.revertError, err.Error())
			}
		})
	}
}
