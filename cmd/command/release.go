package command

import (
	"github.com/mitchellh/cli"
)

const (
	releaseSynopsis = `Creates release`
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
func NewRelease(ui cli.Ui) *Release {
	return &Release{
		ui: ui,
	}
}

// Synopsis returns the short one-line synopsis of the command.
func (b *Release) Synopsis() string {
	return releaseSynopsis
}

// Help returns the long help text including usage, description, and list of flags for the command
func (b *Release) Help() string {
	return releaseHelp
}

// Run runs the actual command with the given CLI instance and command-line arguments
func (b *Release) Run(args []string) int {
	b.ui.Output("release command run")
	return 0
}
