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
		ui            cli.Ui
		spec          *spec.Spec
		workDir       string
		githubToken   string
		expectedError string
	}{
		{
			name:        "Success",
			ui:          &mockUI{},
			spec:        &spec.Spec{},
			workDir:     ".",
			githubToken: "github-token",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f, err := New(tc.ui, tc.spec, tc.workDir, tc.githubToken)

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
		ui      cli.Ui
		message string
	}{
		{
			ui:      &mockUI{},
			message: "Hello, World!",
		},
	}

	for _, tc := range tests {
		ui := &mockUI{}
		f := &formula{
			Ui: ui,
		}

		t.Run("Info", func(t *testing.T) {
			f.Info(tc.message)
			assert.Equal(t, tc.message, ui.InfoInMessage)
		})

		t.Run("Warn", func(t *testing.T) {
			f.Warn(tc.message)
			assert.Equal(t, tc.message, ui.WarnInMessage)
		})

		t.Run("Error", func(t *testing.T) {
			f.Error(tc.message)
			assert.Equal(t, tc.message, ui.ErrorInMessage)
		})
	}
}
