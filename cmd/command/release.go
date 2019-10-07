package command

import (
	"context"
	"errors"
	"time"

	"github.com/mitchellh/cli"
	"github.com/moorara/cherry/internal/action"
	"github.com/moorara/cherry/internal/spec"
	"github.com/moorara/cherry/pkg/cui"
	"github.com/moorara/cherry/pkg/semver"
)

const (
	releaseErr     = 30
	releaseFlagErr = 31

	releaseTimeout = 10 * time.Minute

	releaseSynopsis = `create a new release`
	releaseHelp     = `
	Use this command for creating a new release.

	Flags:

		-patch:    create a patch version release                       (default: true)
		-minor:    create a minor version release                       (default: false)
		-major:    create a major version release                       (default: false)
		-comment:  add a comment for the release
		-model     release model: master, branch                        (default: master)
		-build:    build the artifacts and include them in the release  (default: false)
	
	Examples:

		cherry release
		cherry release -build
		cherry release -minor
		cherry release -minor -build
		cherry release -major
		cherry release -major -build
		cherry release -comment "release comment"
	`
)

// release is the release command.
type release struct {
	ui     cui.CUI
	Spec   *spec.Spec // This needs to be a pointer, so updates made by flag.Parse will be available to downstream consumers
	action action.Action
}

// NewRelease creates a new release command.
func NewRelease(ui cui.CUI, workDir, githubToken string, s *spec.Spec) (cli.Command, error) {
	if githubToken == "" {
		return nil, errors.New("github token is not set")
	}

	return &release{
		ui:     ui,
		Spec:   s,
		action: action.NewRelease(ui, workDir, githubToken, s),
	}, nil
}

// Synopsis returns a short one-line synopsis of the command.
func (c *release) Synopsis() string {
	return releaseSynopsis
}

// Help returns a long help text including usage, description, and list of flags for the command.
func (c *release) Help() string {
	return releaseHelp
}

// Run runs the actual command with the given command-line arguments.
func (c *release) Run(args []string) int {
	var segment semver.Segment
	var patch, minor, major bool
	var comment string

	fs := c.Spec.Release.FlagSet()
	fs.BoolVar(&patch, "patch", true, "")
	fs.BoolVar(&minor, "minor", false, "")
	fs.BoolVar(&major, "major", false, "")
	fs.StringVar(&comment, "comment", "", "")
	fs.Usage = func() {
		c.ui.Outputf(c.Help())
	}

	if err := fs.Parse(args); err != nil {
		return releaseFlagErr
	}

	// Patch default is true
	if patch {
		segment = semver.Patch
	}

	// Minor is preferred over patch
	if minor {
		segment = semver.Minor
	}

	// Major is preferred over minor and patch
	if major {
		segment = semver.Major
	}

	ctx := context.Background()
	ctx = action.ContextForRelease(ctx, segment, comment)
	ctx, cancel := context.WithTimeout(ctx, releaseTimeout)
	defer cancel()

	if err := c.action.Run(ctx); err != nil {
		c.ui.Errorf("%s", err)
		return releaseErr
	}

	return 0
}
