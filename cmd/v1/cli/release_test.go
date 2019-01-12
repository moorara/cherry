package cli

import (
	"errors"
	"testing"

	"github.com/mitchellh/cli"
	"github.com/moorara/cherry/internal/v1/formula"
	"github.com/moorara/cherry/internal/v1/spec"
	"github.com/stretchr/testify/assert"
)

func TestNewRelease(t *testing.T) {
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
			cmd, err := NewRelease(tc.ui, tc.spec, tc.formula)

			assert.NoError(t, err)
			assert.NotNil(t, cmd)
			assert.Equal(t, tc.ui, cmd.Ui)
			assert.Equal(t, tc.spec, cmd.Spec)
			assert.Equal(t, tc.formula, cmd.Formula)
		})
	}
}

func TestReleaseSynopsis(t *testing.T) {
	cmd := &Release{}
	synopsis := cmd.Synopsis()
	assert.Equal(t, releaseSynopsis, synopsis)
}

func TestReleaseHelp(t *testing.T) {
	cmd := &Release{}
	help := cmd.Help()
	assert.Equal(t, releaseHelp, help)
}

func TestReleaseRun(t *testing.T) {
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
			expectedExit: releaseFlagError,
		},
		{
			name: "ReleaseFails",
			ui:   &mockUI{},
			spec: &spec.Spec{},
			formula: &mockFormula{
				ReleaseOutError: errors.New("release error"),
			},
			args:         []string{},
			expectedExit: releaseError,
		},
		{
			name:         "PatchReleaseSuccess",
			ui:           &mockUI{},
			spec:         &spec.Spec{},
			formula:      &mockFormula{},
			args:         []string{"-patch"},
			expectedExit: 0,
		},
		{
			name:         "MinorReleaseSuccess",
			ui:           &mockUI{},
			spec:         &spec.Spec{},
			formula:      &mockFormula{},
			args:         []string{"-minor"},
			expectedExit: 0,
		},
		{
			name:         "MajorReleaseSuccess",
			ui:           &mockUI{},
			spec:         &spec.Spec{},
			formula:      &mockFormula{},
			args:         []string{"-major"},
			expectedExit: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd := &Release{
				Ui:      tc.ui,
				Spec:    tc.spec,
				Formula: tc.formula,
			}

			exit := cmd.Run(tc.args)

			assert.Equal(t, tc.expectedExit, exit)
		})
	}
}
