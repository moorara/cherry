package command

import (
	"bytes"
	"errors"
	"testing"
	"text/template"

	"github.com/mitchellh/cli"
	"github.com/moorara/cherry/internal/spec"
	"github.com/moorara/cherry/pkg/cui"
	"github.com/stretchr/testify/assert"
)

func TestNewBuild(t *testing.T) {
	tests := []struct {
		name          string
		ui            cui.CUI
		workDir       string
		spec          spec.Spec
		expectedError error
	}{
		{
			name:    "OK",
			ui:      &mockCUI{},
			workDir: ".",
			spec:    spec.Spec{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd, err := NewBuild(tc.ui, tc.workDir, tc.spec)

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

func TestBuildSynopsis(t *testing.T) {
	cmd := &build{}
	synopsis := cmd.Synopsis()

	assert.Equal(t, buildSynopsis, synopsis)
}

func TestBuildHelp(t *testing.T) {
	tests := []struct {
		name string
		cmd  cli.Command
	}{
		{
			name: "OK",
			cmd: &build{
				Spec: spec.Spec{
					Build: spec.Build{
						CrossCompile:   true,
						MainFile:       "main.go",
						BinaryFile:     "bin/cherry",
						VersionPackage: "cmd/version",
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			tmpl := template.Must(template.New("help").Parse(buildHelp))
			_ = tmpl.Execute(&buf, tc.cmd)
			expectedHelp := buf.String()

			help := tc.cmd.Help()
			assert.Equal(t, expectedHelp, help)
		})
	}
}

func TestBuildRun(t *testing.T) {
	tests := []struct {
		name         string
		cmd          cli.Command
		args         []string
		expectedExit int
	}{

		{
			name: "InvalidFlags",
			cmd: &build{
				ui:   &mockCUI{},
				Spec: spec.Spec{},
			},
			args:         []string{"-unknown"},
			expectedExit: buildFlagErr,
		},
		{
			name: "ActionFails",
			cmd: &build{
				ui:   &mockCUI{},
				Spec: spec.Spec{},
				action: &mockAction{
					RunOutError: errors.New("error on run: action"),
				},
			},
			args:         []string{},
			expectedExit: buildErr,
		},
		{
			name: "Success",
			cmd: &build{
				ui:     &mockCUI{},
				Spec:   spec.Spec{},
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
