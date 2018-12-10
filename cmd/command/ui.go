package command

import (
	"errors"
	"os"

	"github.com/mitchellh/cli"
)

type (
	// Logger is the interface for the logger struct
	Logger interface {
		Debug(...interface{}) error
		Info(...interface{}) error
		Warn(...interface{}) error
		Error(...interface{}) error
	}

	// UI is a CLI UI
	UI struct {
		cli.Ui
	}

	// LoggerUI is a CLI UI for logging outputs
	LoggerUI struct {
		logger Logger
	}
)

// NewUI creates a new CLI UI
func NewUI() *UI {
	return &UI{
		Ui: &cli.BasicUi{
			Reader:      os.Stdin,
			Writer:      os.Stdout,
			ErrorWriter: os.Stderr,
		},
	}
}

// Colored creates a new UI with colored outputs
func (u *UI) Colored() *UI {
	return &UI{
		Ui: &cli.ColoredUi{
			Ui:          u.Ui,
			OutputColor: cli.UiColorNone,
			InfoColor:   cli.UiColorGreen,
			ErrorColor:  cli.UiColorRed,
			WarnColor:   cli.UiColorYellow,
		},
	}
}

// Concurrent creates a new UI which is safe to be used by multiple goroutines
func (u *UI) Concurrent() *UI {
	return &UI{
		Ui: &cli.ConcurrentUi{
			Ui: u.Ui,
		},
	}
}

// NewLoggerUI creates new CLI UI which uses a logger for printing outputs
func NewLoggerUI(logger Logger) *LoggerUI {
	return &LoggerUI{
		logger: logger,
	}
}

// Output implements cli.Ui Output method
func (u *LoggerUI) Output(message string) {
	u.logger.Info("message", message)
}

// Info implements cli.Ui Info method
func (u *LoggerUI) Info(message string) {
	u.logger.Info("message", message)
}

// Warn implements cli.Ui Warn method
func (u *LoggerUI) Warn(message string) {
	u.logger.Warn("message", message)
}

// Error implements cli.Ui Error method
func (u *LoggerUI) Error(message string) {
	u.logger.Error("message", message)
}

// Ask implements cli.Ui Ask method
func (u *LoggerUI) Ask(query string) (string, error) {
	return "", errors.New("logger ui does not support input")
}

// AskSecret implements cli.Ui AskSecret method
func (u *LoggerUI) AskSecret(query string) (string, error) {
	return "", errors.New("logger ui does not support input")
}
