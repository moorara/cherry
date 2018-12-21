package cli

import (
	"context"

	"github.com/mitchellh/cli"
	"github.com/moorara/cherry/internal/v1/formula"
	"github.com/moorara/cherry/internal/v1/spec"
)

type (
	// Release is the release CLI command
	Release struct {
		cli.Ui
		spec.Spec
		WorkDir     string
		GithubToken string
	}
)

const (
	releaseError     = 40
	releaseFlagError = 41
	releaseSynopsis  = `Create release`
	releaseHelp      = `
	Use this command for creating a new release.

	Flags:

		-patch:    create a patch version release                       (default: true)
		-minor:    create a minor version release                       (default: false)
		-major:    create a major version release                       (default: false)
		-comment:  add a comment for the release
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

// NewRelease create a new release command
func NewRelease(ui cli.Ui, spec spec.Spec, workDir, githubToken string) (*Release, error) {
	cmd := &Release{
		Ui:          ui,
		Spec:        spec,
		WorkDir:     workDir,
		GithubToken: githubToken,
	}

	return cmd, nil
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
	var patch, minor, major bool
	var comment string

	// Parse command flags
	fs := c.Spec.Release.FlagSet()
	fs.BoolVar(&patch, "patch", true, "")
	fs.BoolVar(&minor, "minor", false, "")
	fs.BoolVar(&major, "major", false, "")
	fs.StringVar(&comment, "comment", "", "")
	fs.Usage = func() { c.Ui.Output(c.Help()) }
	if err := fs.Parse(args); err != nil {
		return releaseFlagError
	}

	release, err := formula.NewRelease(c.Ui, c.Spec, c.WorkDir, c.GithubToken)
	if err != nil {
		c.Ui.Error(err.Error())
		return buildError
	}

	var level formula.ReleaseLevel
	if patch {
		level = formula.PatchRelease
	} else if minor {
		level = formula.MinorRelease
	} else if major {
		level = formula.MajorRelease
	}

	ctx, cancel := context.WithTimeout(context.Background(), buildTimeout)
	defer cancel()

	err = release.Release(ctx, level, comment)
	if err != nil {
		c.Ui.Error(err.Error())
		return buildError
	}

	return 0
}
