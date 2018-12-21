package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockUI struct {
	OutputInMessage    string
	InfoInMessage      string
	WarnInMessage      string
	ErrorInMessage     string
	AskInQuery         string
	AskOutString       string
	AskOutError        error
	AskSecretInQuery   string
	AskSecretOutString string
	AskSecretOutError  error
}

func (m *mockUI) Output(message string) {
	m.OutputInMessage = message
}

func (m *mockUI) Info(message string) {
	m.InfoInMessage = message
}

func (m *mockUI) Warn(message string) {
	m.WarnInMessage = message
}

func (m *mockUI) Error(message string) {
	m.ErrorInMessage = message
}

func (m *mockUI) Ask(query string) (string, error) {
	m.AskInQuery = query
	return m.AskOutString, m.AskOutError
}

func (m *mockUI) AskSecret(query string) (string, error) {
	m.AskSecretInQuery = query
	return m.AskSecretOutString, m.AskSecretOutError
}

type mockLogger struct {
	DebugInKV     []interface{}
	DebugOutError error
	InfoInKV      []interface{}
	InfoOutError  error
	WarnInKV      []interface{}
	WarnOutError  error
	ErrorInKV     []interface{}
	ErrorOutError error
}

func (m *mockLogger) Debug(kv ...interface{}) error {
	m.DebugInKV = kv
	return m.DebugOutError
}

func (m *mockLogger) Info(kv ...interface{}) error {
	m.InfoInKV = kv
	return m.InfoOutError
}

func (m *mockLogger) Warn(kv ...interface{}) error {
	m.WarnInKV = kv
	return m.WarnOutError
}

func (m *mockLogger) Error(kv ...interface{}) error {
	m.ErrorInKV = kv
	return m.ErrorOutError
}

func TestUI(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"OK"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var ui *UI

			t.Run("NewUI", func(t *testing.T) {
				ui = NewUI()
				assert.NotNil(t, ui)
			})

			t.Run("Colored", func(t *testing.T) {
				ui = ui.Colored()
				assert.NotNil(t, ui)
			})

			t.Run("Concurrent", func(t *testing.T) {
				ui = ui.Concurrent()
				assert.NotNil(t, ui)
			})
		})
	}
}

func TestLoggerUI(t *testing.T) {
	tests := []struct {
		name          string
		outputMessage string
		infoMessage   string
		warnMessage   string
		errorMessage  string
		askQuery      string
		secretQuery   string
	}{
		{"OK", "", "", "output message", "info message", "warn message", "error message"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var ui *LoggerUI
			logger := &mockLogger{}

			t.Run("NewLoggerUI", func(t *testing.T) {
				ui = NewLoggerUI(logger)
				assert.NotNil(t, ui)
			})

			t.Run("Output", func(t *testing.T) {
				ui.Output(tc.outputMessage)
				assert.Equal(t, "message", logger.InfoInKV[0])
				assert.Equal(t, tc.outputMessage, logger.InfoInKV[1])
			})

			t.Run("Info", func(t *testing.T) {
				ui.Info(tc.infoMessage)
				assert.Equal(t, "message", logger.InfoInKV[0])
				assert.Equal(t, tc.infoMessage, logger.InfoInKV[1])
			})

			t.Run("Warn", func(t *testing.T) {
				ui.Warn(tc.warnMessage)
				assert.Equal(t, "message", logger.WarnInKV[0])
				assert.Equal(t, tc.warnMessage, logger.WarnInKV[1])
			})

			t.Run("Error", func(t *testing.T) {
				ui.Error(tc.errorMessage)
				assert.Equal(t, "message", logger.ErrorInKV[0])
				assert.Equal(t, tc.errorMessage, logger.ErrorInKV[1])
			})

			t.Run("Ask", func(t *testing.T) {
				in, err := ui.Ask(tc.askQuery)
				assert.Error(t, err)
				assert.Empty(t, in)
			})

			t.Run("AskSecret", func(t *testing.T) {
				secret, err := ui.AskSecret(tc.secretQuery)
				assert.Error(t, err)
				assert.Empty(t, secret)
			})
		})
	}
}
