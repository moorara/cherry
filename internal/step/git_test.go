package step

import (
	"context"
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

func TestGitStatusDry(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		expectedError string
	}{
		{
			name:          "Error",
			workDir:       os.TempDir(),
			expectedError: `exit status 128: fatal: not a git repository (or any of the parent directories): .git`,
		},
		{
			name:    "Success",
			workDir: ".",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitStatus{
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

func TestGitStatusRun(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		expectedError string
	}{
		{
			name:          "Error",
			workDir:       os.TempDir(),
			expectedError: `exit status 128: fatal: not a git repository (or any of the parent directories): .git`,
		},
		{
			name:    "Success",
			workDir: ".",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitStatus{
				WorkDir: tc.workDir,
			}

			ctx := context.Background()
			err := step.Run(ctx)

			if tc.expectedError == "" {
				assert.NoError(t, err)
				// TODO: Test results (IsClean)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}

func TestGitStatusRevert(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		expectedError string
	}{
		{
			name:    "Success",
			workDir: ".",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitStatus{
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

func TestGitGetRepoDry(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		expectedError string
	}{
		{
			name:          "Error",
			workDir:       os.TempDir(),
			expectedError: `exit status 128: fatal: not a git repository (or any of the parent directories): .git`,
		},
		{
			name:    "Success",
			workDir: ".",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitGetRepo{
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

func TestGitGetRepoRun(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		expectedError string
	}{
		{
			name:          "Error",
			workDir:       os.TempDir(),
			expectedError: `exit status 128: fatal: not a git repository (or any of the parent directories): .git`,
		},
		{
			name:    "Success",
			workDir: ".",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitGetRepo{
				WorkDir: tc.workDir,
			}

			ctx := context.Background()
			err := step.Run(ctx)

			if tc.expectedError == "" {
				assert.NoError(t, err)
				assert.NotEmpty(t, step.Result.Owner)
				assert.NotEmpty(t, step.Result.Name)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}

func TestGitGetRepoRevert(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		expectedError string
	}{
		{
			name:    "Success",
			workDir: ".",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitGetRepo{
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

func TestGitGetBranchDry(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		expectedError string
	}{
		{
			name:          "Error",
			workDir:       os.TempDir(),
			expectedError: `exit status 128: fatal: not a git repository (or any of the parent directories): .git`,
		},
		{
			name:    "Success",
			workDir: ".",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitGetBranch{
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

func TestGitGetBranchRun(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		expectedError string
	}{
		{
			name:          "Error",
			workDir:       os.TempDir(),
			expectedError: `exit status 128: fatal: not a git repository (or any of the parent directories): .git`,
		},
		{
			name:    "Success",
			workDir: ".",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitGetBranch{
				WorkDir: tc.workDir,
			}

			ctx := context.Background()
			err := step.Run(ctx)

			if tc.expectedError == "" {
				assert.NoError(t, err)
				assert.NotEmpty(t, step.Result.Name)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}

func TestGitGetBranchRevert(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		expectedError string
	}{
		{
			name:    "Success",
			workDir: ".",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitGetBranch{
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

func TestGitGetHEADDry(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		expectedError string
	}{
		{
			name:          "Error",
			workDir:       os.TempDir(),
			expectedError: `exit status 128: fatal: not a git repository (or any of the parent directories): .git`,
		},
		{
			name:    "FullSHA",
			workDir: ".",
		},
		{
			name:    "ShortSHA",
			workDir: ".",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitGetHEAD{
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

func TestGitGetHEADRun(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		expectedError string
	}{
		{
			name:          "Error",
			workDir:       os.TempDir(),
			expectedError: `exit status 128: fatal: not a git repository (or any of the parent directories): .git`,
		},
		{
			name:    "FullSHA",
			workDir: ".",
		},
		{
			name:    "ShortSHA",
			workDir: ".",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitGetHEAD{
				WorkDir: tc.workDir,
			}

			ctx := context.Background()
			err := step.Run(ctx)

			if tc.expectedError == "" {
				assert.NoError(t, err)
				assert.Len(t, step.Result.SHA, 40)
				assert.Len(t, step.Result.ShortSHA, 7)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}

func TestGitGetHEADRevert(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		expectedError string
	}{
		{
			name:    "Success",
			workDir: ".",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitGetHEAD{
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

func TestGitAddDry(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		files         []string
		expectedError string
	}{
		{
			name:          "Error",
			workDir:       os.TempDir(),
			files:         []string{"."},
			expectedError: `exit status 128: fatal: not a git repository (or any of the parent directories): .git`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitAdd{
				WorkDir: tc.workDir,
				Files:   tc.files,
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

func TestGitAddRun(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		files         []string
		expectedError string
	}{
		{
			name:          "Error",
			workDir:       os.TempDir(),
			files:         []string{"."},
			expectedError: `exit status 128: fatal: not a git repository (or any of the parent directories): .git`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitAdd{
				WorkDir: tc.workDir,
				Files:   tc.files,
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

func TestGitAddRevert(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		files         []string
		expectedError string
	}{
		{
			name:          "Error",
			workDir:       os.TempDir(),
			files:         []string{"."},
			expectedError: `exit status 128: fatal: not a git repository (or any of the parent directories): .git`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitAdd{
				WorkDir: tc.workDir,
				Files:   tc.files,
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

func TestGitCommitDry(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		message       string
		expectedError string
	}{
		{
			name:          "Error",
			workDir:       os.TempDir(),
			message:       "test message",
			expectedError: `exit status 128: fatal: not a git repository (or any of the parent directories): .git`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitCommit{
				WorkDir: tc.workDir,
				Message: tc.message,
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

func TestGitCommitRun(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		message       string
		expectedError string
	}{
		{
			name:          "Error",
			workDir:       os.TempDir(),
			message:       "test message",
			expectedError: `exit status 128: fatal: not a git repository (or any of the parent directories): .git`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitCommit{
				WorkDir: tc.workDir,
				Message: tc.message,
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

func TestGitCommitRevert(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		message       string
		expectedError string
	}{
		{
			name:          "Error",
			workDir:       os.TempDir(),
			message:       "test message",
			expectedError: `exit status 128: fatal: not a git repository (or any of the parent directories): .git`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitCommit{
				WorkDir: tc.workDir,
				Message: tc.message,
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

func TestGitTagDry(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		tag           string
		annotation    string
		expectedError string
	}{
		{
			name:          "Error",
			workDir:       os.TempDir(),
			tag:           "test-tag",
			annotation:    "",
			expectedError: "exit status 128: fatal: not a git repository (or any of the parent directories): .git",
		},
		{
			name:          "Error",
			workDir:       os.TempDir(),
			tag:           "test-tag",
			annotation:    "annotation message",
			expectedError: `exit status 128: fatal: not a git repository (or any of the parent directories): .git`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitTag{
				WorkDir:    tc.workDir,
				Tag:        tc.tag,
				Annotation: tc.annotation,
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

func TestGitTagRun(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		tag           string
		annotation    string
		expectedError string
	}{
		{
			name:          "Error",
			workDir:       os.TempDir(),
			tag:           "test-tag",
			annotation:    "",
			expectedError: "exit status 128: fatal: not a git repository (or any of the parent directories): .git",
		},
		{
			name:          "Error",
			workDir:       os.TempDir(),
			tag:           "test-tag",
			annotation:    "annotation message",
			expectedError: `exit status 128: fatal: not a git repository (or any of the parent directories): .git`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitTag{
				WorkDir:    tc.workDir,
				Tag:        tc.tag,
				Annotation: tc.annotation,
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

func TestGitTagRevert(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		tag           string
		annotation    string
		expectedError string
	}{
		{
			name:          "Error",
			workDir:       os.TempDir(),
			tag:           "test-tag",
			annotation:    "",
			expectedError: "exit status 128: fatal: not a git repository (or any of the parent directories): .git",
		},
		{
			name:          "Error",
			workDir:       os.TempDir(),
			tag:           "test-tag",
			annotation:    "annotation message",
			expectedError: `exit status 128: fatal: not a git repository (or any of the parent directories): .git`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitTag{
				WorkDir:    tc.workDir,
				Tag:        tc.tag,
				Annotation: tc.annotation,
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

func TestGitPushDry(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		expectedError string
	}{
		{
			name:          "Error",
			workDir:       os.TempDir(),
			expectedError: `exit status 128: fatal: not a git repository (or any of the parent directories): .git`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitPush{
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

func TestGitPushRun(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		expectedError string
	}{
		{
			name:          "Error",
			workDir:       os.TempDir(),
			expectedError: `exit status 128: fatal: not a git repository (or any of the parent directories): .git`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitPush{
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

func TestGitPushRevert(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		expectedError string
	}{
		{
			name:          "Error",
			workDir:       os.TempDir(),
			expectedError: `cannot revert git push`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitPush{
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

func TestGitPushTagDry(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		tag           string
		expectedError string
	}{
		{
			name:          "Error",
			workDir:       os.TempDir(),
			tag:           "v0.1.0",
			expectedError: `exit status 128: fatal: not a git repository (or any of the parent directories): .git`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitPushTag{
				WorkDir: tc.workDir,
				Tag:     tc.tag,
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

func TestGitPushTagRun(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		tag           string
		expectedError string
	}{
		{
			name:          "Error",
			workDir:       os.TempDir(),
			tag:           "v0.1.0",
			expectedError: `exit status 128: fatal: not a git repository (or any of the parent directories): .git`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitPushTag{
				WorkDir: tc.workDir,
				Tag:     tc.tag,
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

func TestGitPushTagRevert(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		tag           string
		expectedError string
	}{
		{
			name:          "Error",
			workDir:       os.TempDir(),
			tag:           "v0.1.0",
			expectedError: `cannot revert git push`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitPushTag{
				WorkDir: tc.workDir,
				Tag:     tc.tag,
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

func TestGitPullDry(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		expectedError string
	}{
		{
			name:          "Error",
			workDir:       os.TempDir(),
			expectedError: `exit status 128: fatal: not a git repository (or any of the parent directories): .git`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitPull{
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

func TestGitPullRun(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		expectedError string
	}{
		{
			name:          "Error",
			workDir:       os.TempDir(),
			expectedError: `exit status 128: fatal: not a git repository (or any of the parent directories): .git`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitPull{
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

func TestGitPullRevert(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		expectedError string
	}{
		{
			name:          "Error",
			workDir:       os.TempDir(),
			expectedError: `cannot revert git pull`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := GitPull{
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
