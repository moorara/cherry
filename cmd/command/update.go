package command

import (
	"context"
	"flag"
	"time"

	"github.com/mitchellh/cli"
	"github.com/moorara/cherry/internal/action"
	"github.com/moorara/cherry/pkg/cui"
)

const (
	updateFlagErr   = 41
	updateDryErr    = 42
	updateRunErr    = 43
	updateRevertErr = 44

	updateTimeout = time.Minute

	updateSynopsis = `update cherry`
	updateHelp     = `
	Use this command for updating cherry to the latest release.
	
	Examples:

		cherry update
	`
)

// update is the update command.
type update struct {
	ui     cui.CUI
	action action.Action
}

// NewUpdate creates a new update command.
func NewUpdate(ui cui.CUI, githubToken string) (cli.Command, error) {
	return &update{
		ui:     ui,
		action: action.NewUpdate(ui, githubToken),
	}, nil
}

// Synopsis returns a short one-line synopsis of the command.
func (c *update) Synopsis() string {
	return updateSynopsis
}

// Help returns a long help text including usage, description, and list of flags for the command.
func (c *update) Help() string {
	return updateHelp
}

// Run runs the actual command with the given command-line arguments.
func (c *update) Run(args []string) int {
	fs := flag.NewFlagSet("update", flag.ContinueOnError)
	fs.Usage = func() {
		c.ui.Outputf(c.Help())
	}

	if err := fs.Parse(args); err != nil {
		return updateFlagErr
	}

	ctx, cancel := context.WithTimeout(context.Background(), updateTimeout)
	defer cancel()

	// Try finding any possible failure before running the command
	if err := c.action.Dry(ctx); err != nil {
		c.ui.Errorf("%s", err)
		return updateDryErr
	}

	// Running the command
	if err := c.action.Run(ctx); err != nil {
		c.ui.Errorf("%s", err)

		// Try reverting back any side effect in case of failure
		if err := c.action.Revert(ctx); err != nil {
			c.ui.Errorf("%s", err)
			return updateRevertErr
		}

		return updateRunErr
	}

	return 0
}
