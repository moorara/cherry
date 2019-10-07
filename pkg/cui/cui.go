package cui

import (
	"fmt"
	"os"

	"github.com/mitchellh/cli"
)

// CUI is the command-line user interface.
type CUI interface {
	Outputf(string, ...interface{})
	Infof(string, ...interface{})
	Warnf(string, ...interface{})
	Errorf(string, ...interface{})
}

type cui struct {
	ui cli.Ui
}

// New creates a new instance of CUI which is colored and concurrency safe.
func New() CUI {
	return &cui{
		ui: &cli.ConcurrentUi{
			Ui: &cli.ColoredUi{
				Ui: &cli.BasicUi{
					Reader:      os.Stdin,
					Writer:      os.Stdout,
					ErrorWriter: os.Stderr,
				},
				OutputColor: cli.UiColorNone,
				InfoColor:   cli.UiColorGreen,
				ErrorColor:  cli.UiColorRed,
				WarnColor:   cli.UiColorYellow,
			},
		},
	}
}

func (c *cui) Outputf(format string, v ...interface{}) {
	if c.ui != nil {
		c.ui.Output(fmt.Sprintf(format, v...))
	}
}

func (c *cui) Infof(format string, v ...interface{}) {
	if c.ui != nil {
		c.ui.Info(fmt.Sprintf(format, v...))
	}
}

func (c *cui) Warnf(format string, v ...interface{}) {
	if c.ui != nil {
		c.ui.Warn(fmt.Sprintf(format, v...))
	}
}

func (c *cui) Errorf(format string, v ...interface{}) {
	if c.ui != nil {
		c.ui.Error(fmt.Sprintf(format, v...))
	}
}
