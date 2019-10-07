package cui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockUI struct {
	OutputInMsg string
	InfoInMsg   string
	WarnInMsg   string
	ErrorInMsg  string

	AskInQuery   string
	AskOutResult string
	AskOutError  error

	AskSecretInQuery   string
	AskSecretOutResult string
	AskSecretOutError  error
}

func (m *mockUI) Output(msg string) {
	m.OutputInMsg = msg
}

func (m *mockUI) Info(msg string) {
	m.InfoInMsg = msg
}

func (m *mockUI) Warn(msg string) {
	m.WarnInMsg = msg
}

func (m *mockUI) Error(msg string) {
	m.ErrorInMsg = msg
}

func (m *mockUI) Ask(query string) (string, error) {
	m.AskInQuery = query
	return m.AskOutResult, m.AskOutError
}

func (m *mockUI) AskSecret(query string) (string, error) {
	m.AskSecretInQuery = query
	return m.AskSecretOutResult, m.AskSecretOutError
}

func TestNew(t *testing.T) {
	ci := New()
	assert.NotNil(t, ci)

	cs, ok := ci.(*cui)
	assert.True(t, ok)
	assert.NotNil(t, cs)
	assert.NotNil(t, cs.ui)
}

func TestCUI(t *testing.T) {
	tests := []struct {
		ui             *mockUI
		format         string
		vals           []interface{}
		expectedOutput string
	}{
		{
			ui:     &mockUI{},
			format: "Hello, %s!",
			vals: []interface{}{
				"World",
			},
			expectedOutput: "Hello, World!",
		},
	}

	for _, tc := range tests {
		c := &cui{
			ui: tc.ui,
		}

		t.Run("Outputf", func(t *testing.T) {
			c.Outputf(tc.format, tc.vals...)
			assert.Equal(t, tc.expectedOutput, tc.ui.OutputInMsg)
		})

		t.Run("Infof", func(t *testing.T) {
			c.Infof(tc.format, tc.vals...)
			assert.Equal(t, tc.expectedOutput, tc.ui.InfoInMsg)
		})

		t.Run("Warnf", func(t *testing.T) {
			c.Warnf(tc.format, tc.vals...)
			assert.Equal(t, tc.expectedOutput, tc.ui.WarnInMsg)
		})

		t.Run("Errorf", func(t *testing.T) {
			c.Errorf(tc.format, tc.vals...)
			assert.Equal(t, tc.expectedOutput, tc.ui.ErrorInMsg)
		})
	}
}
