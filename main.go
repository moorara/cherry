package main

import (
	"fmt"
	"os"

	"github.com/mitchellh/cli"
	"github.com/moorara/cherry/internal/command"
	"github.com/moorara/cherry/internal/spec"
	"github.com/moorara/cherry/version"
)

const (
	specErr = 101
)

func main() {
	ui := &cli.ConcurrentUi{
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
	}

	// Read the spec from file if any
	s, err := spec.FromFile()
	if err != nil {
		ui.Error(fmt.Sprintf("Error on reading spec file: %s", err))
		os.Exit(specErr)
	}

	// Get default values for zero fields
	s = s.WithDefaults()
	s.ToolVersion = version.Version

	c := cli.NewCLI("cherry", version.String())
	c.Args = os.Args[1:]
	c.Commands = map[string]cli.CommandFactory{
		"build": func() (cli.Command, error) {
			return command.NewBuild(ui, s)
		},
		"release": func() (cli.Command, error) {
			return command.NewRelease(ui, s)
		},
		"update": func() (cli.Command, error) {
			return command.NewUpdate(ui)
		},
	}

	code, err := c.Run()
	if err != nil {
		ui.Error(err.Error())
	}

	os.Exit(code)
}
