package cli

import (
	"context"
	"time"

	"github.com/mitchellh/cli"
	"github.com/moorara/cherry/internal/v1/formula"
	"github.com/moorara/cherry/internal/v1/spec"
)

type (
	// Test is the test CLI command
	Test struct {
		cli.Ui
		spec.Spec
		WorkDir string
	}
)

const (
	testError     = 40
	testFlagError = 41
	testTimeout   = 1 * time.Minute
	testSynopsis  = `Run tests`
	testHelp      = `
	Use this command for running unit tests and generating coverage report.
	Currently, this command can only test Go applications.

	Flags:
	
		-report-path:  the path for coverage report  (default: coverage)
	
	Examples:

		cherry test
		cherry test -report-path report
	`
)

// NewTest create a new test command
func NewTest(ui cli.Ui, spec spec.Spec, workDir string) (*Test, error) {
	cmd := &Test{
		Ui:      ui,
		Spec:    spec,
		WorkDir: workDir,
	}

	return cmd, nil
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
	fs := c.Spec.Test.FlagSet()
	fs.Usage = func() { c.Ui.Output(c.Help()) }
	if err := fs.Parse(args); err != nil {
		return testFlagError
	}

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	test := formula.NewTest(c.Ui, c.Spec, c.WorkDir)

	err := test.Cover(ctx)
	if err != nil {
		c.Ui.Error(err.Error())
		return testError
	}

	return 0
}