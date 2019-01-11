package cli

import (
	"errors"
	"testing"

	"github.com/mitchellh/cli"
	"github.com/moorara/cherry/internal/v1/formula"
	"github.com/moorara/cherry/internal/v1/spec"
	"github.com/stretchr/testify/assert"
)

func TestNewTest(t *testing.T) {
	tests := []struct {
		name    string
		ui      cli.Ui
		spec    *spec.Spec
		formula formula.Formula
	}{
		{
			name:    "MockDependencies",
			ui:      &mockUI{},
			spec:    &spec.Spec{},
			formula: &mockFormula{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd, err := NewTest(tc.ui, tc.spec, tc.formula)

			assert.NoError(t, err)
			assert.NotNil(t, cmd)
			assert.Equal(t, tc.ui, cmd.Ui)
			assert.Equal(t, tc.spec, cmd.Spec)
			assert.Equal(t, tc.formula, cmd.Formula)
		})
	}
}

func TestTestSynopsis(t *testing.T) {
	cmd := &Test{}
	synopsis := cmd.Synopsis()
	assert.Equal(t, testSynopsis, synopsis)
}

func TestTestHelp(t *testing.T) {
	cmd := &Test{}
	help := cmd.Help()
	assert.Equal(t, testHelp, help)
}

func TestTestRun(t *testing.T) {
	tests := []struct {
		name         string
		ui           cli.Ui
		spec         *spec.Spec
		formula      formula.Formula
		args         []string
		expectedExit int
	}{
		{
			name:         "ParseFlagFails",
			ui:           &mockUI{},
			spec:         &spec.Spec{},
			formula:      &mockFormula{},
			args:         []string{"-unknown"},
			expectedExit: testFlagError,
		},
		{
			name: "CoverFails",
			ui:   &mockUI{},
			spec: &spec.Spec{},
			formula: &mockFormula{
				CoverOutError: errors.New("test cover error"),
			},
			args:         []string{},
			expectedExit: testError,
		},
		{
			name:         "CoverSuccess",
			ui:           &mockUI{},
			spec:         &spec.Spec{},
			formula:      &mockFormula{},
			args:         []string{},
			expectedExit: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd := &Test{
				Ui:      tc.ui,
				Spec:    tc.spec,
				Formula: tc.formula,
			}

			exit := cmd.Run(tc.args)

			assert.Equal(t, tc.expectedExit, exit)
		})
	}
}
