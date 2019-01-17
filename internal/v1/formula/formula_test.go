package formula

import (
	"testing"

	"github.com/mitchellh/cli"
	"github.com/moorara/cherry/internal/v1/spec"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name          string
		workDir       string
		githubToken   string
		spec          *spec.Spec
		ui            cli.Ui
		expectedError string
	}{
		{
			name:        "Success",
			workDir:     ".",
			githubToken: "github-token",
			spec:        &spec.Spec{},
			ui:          &mockUI{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f, err := New(tc.workDir, tc.githubToken, tc.spec, tc.ui)

			if tc.expectedError != "" {
				assert.Contains(t, err.Error(), tc.expectedError)
				assert.Nil(t, f)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, f)
			}
		})
	}
}

func TestFormula(t *testing.T) {
	tests := []struct {
		ui              cli.Ui
		message         string
		args            []interface{}
		expectedMessage string
	}{
		{
			ui:      &mockUI{},
			message: "Hello, %s!",
			args: []interface{}{
				"World",
			},
			expectedMessage: "Hello, World!",
		},
	}

	for _, tc := range tests {
		ui := &mockUI{}
		f := &formula{
			ui: ui,
		}

		t.Run("Printf", func(t *testing.T) {
			f.Printf(tc.message, tc.args...)
			assert.Equal(t, tc.expectedMessage, ui.OutputInMessage)
		})

		t.Run("Infof", func(t *testing.T) {
			f.Infof(tc.message, tc.args...)
			assert.Equal(t, tc.expectedMessage, ui.InfoInMessage)
		})

		t.Run("Warnf", func(t *testing.T) {
			f.Warnf(tc.message, tc.args...)
			assert.Equal(t, tc.expectedMessage, ui.WarnInMessage)
		})

		t.Run("Errorf", func(t *testing.T) {
			f.Errorf(tc.message, tc.args...)
			assert.Equal(t, tc.expectedMessage, ui.ErrorInMessage)
		})
	}
}
