package cli

import (
	"bytes"
	"errors"
	"testing"
	"text/template"

	"github.com/mitchellh/cli"
	"github.com/moorara/cherry/internal/v1/formula"
	"github.com/moorara/cherry/internal/v1/spec"
	"github.com/stretchr/testify/assert"
)

func TestNewBuild(t *testing.T) {
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
			cmd, err := NewBuild(tc.ui, tc.spec, tc.formula)

			assert.NoError(t, err)
			assert.NotNil(t, cmd)
			assert.Equal(t, tc.ui, cmd.Ui)
			assert.Equal(t, tc.spec, cmd.Spec)
			assert.Equal(t, tc.formula, cmd.Formula)
		})
	}
}

func TestBuildSynopsis(t *testing.T) {
	cmd := &Build{}

	synopsis := cmd.Synopsis()
	assert.Equal(t, buildSynopsis, synopsis)
}

func TestBuildHelp(t *testing.T) {
	tests := []struct {
		spec *spec.Spec
	}{
		{
			spec: &spec.Spec{
				Build: spec.Build{
					CrossCompile:   true,
					MainFile:       "main.go",
					BinaryFile:     "bin/cherry",
					VersionPackage: "cmd/version",
				},
			},
		},
	}

	for _, tc := range tests {
		cmd := &Build{
			Spec: tc.spec,
		}

		var buf bytes.Buffer
		tmpl := template.Must(template.New("help").Parse(buildHelp))
		tmpl.Execute(&buf, cmd)
		expectedHelp := buf.String()

		help := cmd.Help()
		assert.Equal(t, expectedHelp, help)
	}
}

func TestBuildRun(t *testing.T) {
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
			expectedExit: buildFlagError,
		},
		{
			name: "CompileFails",
			ui:   &mockUI{},
			spec: &spec.Spec{
				Build: spec.Build{
					CrossCompile: true,
				},
			},
			formula: &mockFormula{
				CrossCompileOutResult: nil,
				CrossCompileOutError:  errors.New("compile error"),
			},
			args:         []string{},
			expectedExit: buildError,
		},
		{
			name: "CrossCompileFails",
			ui:   &mockUI{},
			spec: &spec.Spec{},
			formula: &mockFormula{
				CompileOutError: errors.New("compile error"),
			},
			args:         []string{},
			expectedExit: buildError,
		},
		{
			name:         "SuccessWithCompile",
			ui:           &mockUI{},
			spec:         &spec.Spec{},
			formula:      &mockFormula{},
			args:         []string{},
			expectedExit: 0,
		},
		{
			name: "SuccessWithCrossCompile",
			ui:   &mockUI{},
			spec: &spec.Spec{
				Build: spec.Build{
					CrossCompile: true,
				},
			},
			formula: &mockFormula{
				CrossCompileOutResult: []string{},
			},
			args:         []string{},
			expectedExit: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd := &Build{
				Ui:      tc.ui,
				Spec:    tc.spec,
				Formula: tc.formula,
			}

			exit := cmd.Run(tc.args)

			assert.Equal(t, tc.expectedExit, exit)
		})
	}
}
