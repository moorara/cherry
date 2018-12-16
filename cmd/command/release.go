package command

import (
	"context"
	"errors"
	"flag"
	"time"

	"github.com/mitchellh/cli"
	"github.com/moorara/cherry/internal/git"
)

const (
	releaseError     = 40
	releaseFlagError = 41
	releaseTimeout   = 1 * time.Minute
	releaseSynopsis  = `Create release`
	releaseHelp      = `
	Use this command for creating a new release.
	`
)

type (
	// Release is the release CLI command
	Release struct {
		ui          cli.Ui
		workDir     string
		githubToken string
		git         git.Git
	}
)

// NewRelease create a new release command
func NewRelease(ui cli.Ui, workDir, githubToken string) (*Release, error) {
	cmd := &Release{
		ui:          ui,
		workDir:     workDir,
		githubToken: githubToken,
		git:         git.New(workDir),
	}

	return cmd, nil
}

func (c *Release) release(ctx context.Context) error {
	clean, err := c.git.IsClean()
	if err != nil {
		return err
	}

	if !clean {
		return errors.New("git repo is not clean and has uncommitted changes")
	}

	return nil
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

	ctx, cancel := context.WithTimeout(context.Background(), buildTimeout)
	defer cancel()

	err := c.release(ctx)
	if err != nil {
		c.ui.Error(err.Error())
		return buildError
	}

	return 0
}
