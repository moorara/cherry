package command

import (
	"flag"
	"time"

	"github.com/mitchellh/cli"
)

const (
	releaseError     = 20
	releaseFlagError = 21

	releaseTimeout = 1 * time.Minute

	releaseSynopsis = `Create release`
	releaseHelp     = `
	Use this command for creating a new release.
	`
)

type (
	// Release is the release CLI command
	Release struct {
		ui cli.Ui
	}
)

// NewRelease create a new release command
func NewRelease(ui cli.Ui) (*Release, error) {
	return &Release{
		ui: ui,
	}, nil
}

// Synopsis returns the short one-line synopsis of the command.
func (c *Release) Synopsis() string {
	return releaseSynopsis
}

// Help returns the long help text including usage, description, and list of flags for the command
func (c *Release) Help() string {
	return releaseHelp
}

// Run runs the actual command with the given CLI instance and command-line arguments
func (c *Release) Run(args []string) int {
	// Parse command flags
	flags := flag.NewFlagSet("release", flag.ContinueOnError)
	flags.Usage = func() { c.ui.Output(c.Help()) }
	if err := flags.Parse(args); err != nil {
		return releaseFlagError
	}

	c.ui.Output("release command run")

	return 0
}
