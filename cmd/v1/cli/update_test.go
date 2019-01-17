package cli

import (
	"errors"
	"testing"

	"github.com/mitchellh/cli"
	"github.com/moorara/cherry/internal/v1/formula"
	"github.com/stretchr/testify/assert"
)

func TestNewUpdate(t *testing.T) {
	tests := []struct {
		name    string
		ui      cli.Ui
		formula formula.Formula
	}{
		{
			name:    "MockDependencies",
			ui:      &mockUI{},
			formula: &mockFormula{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd, err := NewUpdate(tc.ui, tc.formula)

			assert.NoError(t, err)
			assert.NotNil(t, cmd)
			assert.Equal(t, tc.ui, cmd.Ui)
			assert.Equal(t, tc.formula, cmd.Formula)
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
		ui           cli.Ui
		formula      formula.Formula
		args         []string
		expectedExit int
	}{
		{
			name:         "ParseFlagFails",
			ui:           &mockUI{},
			formula:      &mockFormula{},
			args:         []string{"-unknown"},
			expectedExit: updateFlagError,
		},
		{
			name: "UpdateFails",
			ui:   &mockUI{},
			formula: &mockFormula{
				UpdateOutError: errors.New("update error"),
			},
			args:         []string{},
			expectedExit: updateError,
		},
		{
			name:         "CoverSuccess",
			ui:           &mockUI{},
			formula:      &mockFormula{},
			args:         []string{},
			expectedExit: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd := &Update{
				Ui:      tc.ui,
				Formula: tc.formula,
			}

			exit := cmd.Run(tc.args)

			assert.Equal(t, tc.expectedExit, exit)
		})
	}
}
