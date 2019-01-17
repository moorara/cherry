package cli

import (
	"context"
	"flag"
	"os"
	"path/filepath"
	"time"

	"github.com/mitchellh/cli"
	"github.com/moorara/cherry/internal/v1/formula"
)

type (
	// Update is the update cli command
	Update struct {
		cli.Ui
		formula.Formula
	}
)

const (
	updateError        = 50
	updateFlagError    = 51
	updateBinPathError = 52
	updateTimeout      = 2 * time.Minute
	updateSynopsis     = `update cherry`
	updateHelp         = `
	Use this command for updating cherry to the latest release.
	
	Examples:

		cherry update
	`
)

// NewUpdate create a new update command
func NewUpdate(ui cli.Ui, formula formula.Formula) (*Update, error) {
	cmd := &Update{
		Ui:      ui,
		Formula: formula,
	}

	return cmd, nil
}

// Synopsis returns the short one-line synopsis of the command.
func (c *Update) Synopsis() string {
	return updateSynopsis
}

// Help returns the long help text including usage, description, and list of flags for the command
func (c *Update) Help() string {
	return updateHelp
}

// Run runs the actual command with the given command-line arguments
func (c *Update) Run(args []string) int {
	// Parse command flags
	fs := flag.NewFlagSet("update", flag.ContinueOnError)
	fs.Usage = func() { c.Ui.Output(c.Help()) }
	if err := fs.Parse(args); err != nil {
		return updateFlagError
	}

	ctx, cancel := context.WithTimeout(context.Background(), updateTimeout)
	defer cancel()

	binPath, err := filepath.Abs(os.Args[0])
	if err != nil {
		return updateBinPathError
	}

	err = c.Formula.Update(ctx, binPath)
	if err != nil {
		c.Ui.Error(err.Error())
		return updateError
	}

	return 0
}
