package command

import (
	"github.com/mitchellh/cli"
)

const (
	buildSynopsis = `Builds artifacts`
	buildHelp     = `
	Use this command for building artifacts.
	`
)

type (
	// Build is the build CLI command
	Build struct {
		ui cli.Ui
	}
)

// NewBuild create a new build command
func NewBuild(ui cli.Ui) *Build {
	return &Build{
		ui: ui,
	}
}

// Synopsis returns the short one-line synopsis of the command.
func (c *Build) Synopsis() string {
	return buildSynopsis
}

// Help returns the long help text including usage, description, and list of flags for the command
func (c *Build) Help() string {
	return buildHelp
}

// Run runs the actual command with the given CLI instance and command-line arguments
func (c *Build) Run(args []string) int {
	c.ui.Output("build command run")

	return 0
}
