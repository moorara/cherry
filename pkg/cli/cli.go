package cli

import (
	"fmt"
	"os"

	hcli "github.com/mitchellh/cli"
)

// CLI is the command-line interface.
type CLI interface {
	Outputf(string, ...interface{})
	Infof(string, ...interface{})
	Warnf(string, ...interface{})
	Errorf(string, ...interface{})
}

type cli struct {
	ui hcli.Ui
}

// New creates a new instance of CLI which is colored and concurrency safe.
func New() CLI {
	return &cli{
		ui: &hcli.ConcurrentUi{
			Ui: &hcli.ColoredUi{
				Ui: &hcli.BasicUi{
					Reader:      os.Stdin,
					Writer:      os.Stdout,
					ErrorWriter: os.Stderr,
				},
				OutputColor: hcli.UiColorNone,
				InfoColor:   hcli.UiColorGreen,
				ErrorColor:  hcli.UiColorRed,
				WarnColor:   hcli.UiColorYellow,
			},
		},
	}
}

func (c *cli) Outputf(format string, v ...interface{}) {
	if c.ui != nil {
		c.ui.Output(fmt.Sprintf(format, v...))
	}
}

func (c *cli) Infof(format string, v ...interface{}) {
	if c.ui != nil {
		c.ui.Info(fmt.Sprintf(format, v...))
	}
}

func (c *cli) Warnf(format string, v ...interface{}) {
	if c.ui != nil {
		c.ui.Warn(fmt.Sprintf(format, v...))
	}
}

func (c *cli) Errorf(format string, v ...interface{}) {
	if c.ui != nil {
		c.ui.Error(fmt.Sprintf(format, v...))
	}
}
