package command

import (
	"context"
	"flag"
	"os"
	"time"

	"github.com/mitchellh/cli"
	"github.com/moorara/cherry/internal/test"
)

const (
	testError     = 30
	testFlagError = 31

	testTimeout = 1 * time.Minute

	testSynopsis = `Run tests`
	testHelp     = `
	Use this command for running unit tests and generating coverage report.
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
func NewTest(ui cli.Ui) (*Test, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	tester := test.NewTester(wd)

	return &Test{
		ui:     ui,
		tester: tester,
	}, nil
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
	// Parse command flags
	flags := flag.NewFlagSet("test", flag.ContinueOnError)
	flags.Usage = func() { c.ui.Output(c.Help()) }
	if err := flags.Parse(args); err != nil {
		return testFlagError
	}

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	err := c.tester.Cover(ctx)
	if err != nil {
		c.ui.Error(err.Error())
		return testError
	}

	return 0
}
