package command

import (
	"context"
	"time"

	"github.com/mitchellh/cli"
	"github.com/moorara/cherry/internal/test"
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
		ui     cli.Ui
		tester test.Tester
	}
)

// NewTest create a new test command
func NewTest(ui cli.Ui) *Test {
	tester := test.NewTester("./")

	return &Test{
		ui:     ui,
		tester: tester,
	}
}

// Synopsis returns the short one-line synopsis of the command.
func (c *Test) Synopsis() string {
	return testSynopsis
}

// Help returns the long help text including usage, description, and list of flags for the command
func (c *Test) Help() string {
	return testHelp
}

// Run runs the actual command with the given CLI instance and command-line arguments
func (c *Test) Run(args []string) int {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := c.tester.Coverage(ctx)
	if err != nil {
		c.ui.Error(err.Error())
		return 1
	}

	return 0
}
