package command

import (
	"errors"
	"testing"

	"github.com/mitchellh/cli"
	"github.com/moorara/cherry/internal/spec"
	"github.com/moorara/cherry/pkg/cui"
	"github.com/stretchr/testify/assert"
)

func TestNewRelease(t *testing.T) {
	tests := []struct {
		name          string
		ui            cui.CUI
		workDir       string
		githubToken   string
		spec          spec.Spec
		expectedError error
	}{
		{
			name:          "NoGitHubToken",
			ui:            &mockCUI{},
			workDir:       ".",
			githubToken:   "",
			expectedError: errors.New("github token is not set"),
		},
		{
			name:        "OK",
			ui:          &mockCUI{},
			workDir:     ".",
			githubToken: "github-token",
			spec:        spec.Spec{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd, err := NewRelease(tc.ui, tc.workDir, tc.githubToken, tc.spec)

			if tc.expectedError == nil {
				assert.NotNil(t, cmd)
				assert.NoError(t, err)
			} else {
				assert.Nil(t, cmd)
				assert.Equal(t, tc.expectedError, err)
			}
		})
	}
}

func TestReleaseSynopsis(t *testing.T) {
	cmd := &release{}
	synopsis := cmd.Synopsis()

	assert.Equal(t, releaseSynopsis, synopsis)
}

func TestReleaseHelp(t *testing.T) {
	cmd := &release{}
	help := cmd.Help()

	assert.Equal(t, releaseHelp, help)
}

func TestReleaseRun(t *testing.T) {
	tests := []struct {
		name         string
		cmd          cli.Command
		args         []string
		expectedExit int
	}{

		{
			name: "InvalidFlags",
			cmd: &release{
				ui:   &mockCUI{},
				Spec: spec.Spec{},
			},
			args:         []string{"-unknown"},
			expectedExit: releaseFlagErr,
		},
		{
			name: "DryFails",
			cmd: &release{
				ui:   &mockCUI{},
				Spec: spec.Spec{},
				action: &mockAction{
					DryOutError: errors.New("error on dry: action"),
				},
			},
			args:         []string{},
			expectedExit: releaseDryErr,
		},
		{
			name: "RunFails",
			cmd: &release{
				ui:   &mockCUI{},
				Spec: spec.Spec{},
				action: &mockAction{
					RunOutError: errors.New("error on run: action"),
				},
			},
			args:         []string{},
			expectedExit: releaseRunErr,
		},
		{
			name: "RevertFails",
			cmd: &release{
				ui:   &mockCUI{},
				Spec: spec.Spec{},
				action: &mockAction{
					RunOutError:    errors.New("error on run: action"),
					RevertOutError: errors.New("error on revert: action"),
				},
			},
			args:         []string{},
			expectedExit: releaseRevertErr,
		},
		{
			name: "PatchSuccess",
			cmd: &release{
				ui:     &mockCUI{},
				Spec:   spec.Spec{},
				action: &mockAction{},
			},
			args:         []string{"-patch"},
			expectedExit: 0,
		},
		{
			name: "MinorSuccess",
			cmd: &release{
				ui:     &mockCUI{},
				Spec:   spec.Spec{},
				action: &mockAction{},
			},
			args:         []string{"-minor"},
			expectedExit: 0,
		},
		{
			name: "MajorSuccess",
			cmd: &release{
				ui:     &mockCUI{},
				Spec:   spec.Spec{},
				action: &mockAction{},
			},
			args:         []string{"-major"},
			expectedExit: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			exit := tc.cmd.Run(tc.args)

			assert.Equal(t, tc.expectedExit, exit)
		})
	}
}
