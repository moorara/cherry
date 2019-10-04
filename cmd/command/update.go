package command

import (
	"context"
	"flag"
	"time"

	"github.com/moorara/cherry/internal/action"
	"github.com/moorara/cherry/pkg/cui"
)

const (
	updateErr     = 40
	updateFlagErr = 41

	updateTimeout = time.Minute

	updateSynopsis = `update cherry`
	updateHelp     = `
	Use this command for updating cherry to the latest release.
	
	Examples:

		cherry update
	`
)

// Update is the update command.
type Update struct {
	ui     cui.CUI
	action action.Action
}

// NewUpdate creates a new update command.
func NewUpdate(ui cui.CUI, githubToken string) (*Update, error) {
	return &Update{
		ui:     ui,
		action: action.NewUpdate(ui, githubToken),
	}, nil
}

// Synopsis returns a short one-line synopsis of the command.
func (c *Update) Synopsis() string {
	return updateSynopsis
}

// Help returns a long help text including usage, description, and list of flags for the command.
func (c *Update) Help() string {
	return updateHelp
}

// Run runs the actual command with the given command-line arguments.
func (c *Update) Run(args []string) int {
	fs := flag.NewFlagSet("update", flag.ContinueOnError)
	fs.Usage = func() {
		c.ui.Outputf(c.Help())
	}

	err := fs.Parse(args)
	if err != nil {
		return updateFlagErr
	}

	ctx, cancel := context.WithTimeout(context.Background(), updateTimeout)
	defer cancel()

	err = c.action.Run(ctx)
	if err != nil {
		c.ui.Errorf("%s", err)
		return updateErr
	}

	return 0
}
