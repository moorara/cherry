package command

import (
	"github.com/mitchellh/cli"
)

const (
	testSynopsis = `Runs unit tests`
	testHelp     = `
	Use this command for running unit tests.
	`
)

type (
	// Test is the test CLI command
	Test struct {
		ui cli.Ui
	}
)

// NewTest create a new test command
func NewTest(ui cli.Ui) *Test {
	return &Test{
		ui: ui,
	}
}

// Synopsis returns the short one-line synopsis of the command.
func (b *Test) Synopsis() string {
	return testSynopsis
}

// Help returns the long help text including usage, description, and list of flags for the command
func (b *Test) Help() string {
	return testHelp
}

// Run runs the actual command with the given CLI instance and command-line arguments
func (b *Test) Run(args []string) int {
	b.ui.Output("test command run")
	return 0
}
