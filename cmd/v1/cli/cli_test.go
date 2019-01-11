package cli

import (
	"testing"

	"github.com/mitchellh/cli"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name          string
		ui            cli.Ui
		appName       string
		githubToken   string
		expectedError string
	}{
		{
			name:          "Success",
			ui:            &mockUI{},
			appName:       "cherry",
			githubToken:   "github-token",
			expectedError: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			app, err := New(tc.ui, tc.appName, tc.githubToken)

			if tc.expectedError != "" {
				assert.Contains(t, err.Error(), tc.expectedError)
				assert.Nil(t, app)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, app)

				status, err := app.Run()
				assert.NoError(t, err)
				assert.Equal(t, 127, status)
			}
		})
	}
}
