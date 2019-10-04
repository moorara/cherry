package command

import (
	"errors"
	"testing"

	"github.com/moorara/cherry/pkg/cui"
	"github.com/stretchr/testify/assert"
)

func TestNewUpdate(t *testing.T) {
	tests := []struct {
		name          string
		ui            cui.CUI
		githubToken   string
		expectedError error
	}{
		{
			name:        "OK",
			ui:          &mockCUI{},
			githubToken: "github-token",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd, err := NewUpdate(tc.ui, tc.githubToken)

			assert.NotNil(t, cmd)
			assert.Equal(t, tc.expectedError, err)
		})
	}
}

func TestUpdateSynopsis(t *testing.T) {
	cmd := &Update{}
	synopsis := cmd.Synopsis()

	assert.Equal(t, updateSynopsis, synopsis)
}

func TestUpdateHelp(t *testing.T) {
	cmd := &Update{}
	help := cmd.Help()

	assert.Equal(t, updateHelp, help)
}

func TestUpdateRun(t *testing.T) {
	tests := []struct {
		name         string
		cmd          *Update
		args         []string
		expectedExit int
	}{

		{
			name: "InvalidFlags",
			cmd: &Update{
				ui: &mockCUI{},
			},
			args:         []string{"-unknown"},
			expectedExit: updateFlagErr,
		},
		{
			name: "ActionFails",
			cmd: &Update{
				ui: &mockCUI{},
				action: &mockAction{
					RunOutError: errors.New("error on run: action"),
				},
			},
			args:         []string{},
			expectedExit: updateErr,
		},
		{
			name: "Success",
			cmd: &Update{
				ui:     &mockCUI{},
				action: &mockAction{},
			},
			args:         []string{},
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
