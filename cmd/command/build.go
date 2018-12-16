package command

import (
	"context"
	"flag"
	"os"
	"time"

	"github.com/mitchellh/cli"
	"github.com/moorara/cherry/internal/build"
)

const (
	buildError     = 10
	buildFlagError = 11

	buildTimeout = 1 * time.Minute

	buildSynopsis = `Build artifacts`
	buildHelp     = `
	Use this command for building artifacts.

	Flags:
		-main: path to main.go file                (default: main.go)
		-out:  path for binary files               (default: bin/app)
		-all:  build the binary for all platforms  (default: false)
	`
)

type (
	// Build is the build CLI command
	Build struct {
		ui      cli.Ui
		builder build.Builder
	}
)

// NewBuild create a new build command
func NewBuild(ui cli.Ui) (*Build, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	builder := build.NewBuilder(wd)

	return &Build{
		ui:      ui,
		builder: builder,
	}, nil
}

// Synopsis returns the short one-line synopsis of the command.
func (c *Build) Synopsis() string {
	return buildSynopsis
}

// Help returns the long help text including usage, description, and list of flags for the command
func (c *Build) Help() string {
	return buildHelp
}

// Run runs the actual command with the given CLI instance and command-line arguments
func (c *Build) Run(args []string) int {
	var main, out string
	var all bool

	// Parse command flags
	flags := flag.NewFlagSet("build", flag.ContinueOnError)
	flags.StringVar(&main, "main", "main.go", "")
	flags.StringVar(&out, "out", "bin/app", "")
	flags.BoolVar(&all, "all", false, "")
	flags.Usage = func() { c.ui.Output(c.Help()) }
	if err := flags.Parse(args); err != nil {
		return buildFlagError
	}

	ctx, cancel := context.WithTimeout(context.Background(), buildTimeout)
	defer cancel()

	err := c.builder.Build(ctx, main, out)
	if err != nil {
		c.ui.Error(err.Error())
		return buildError
	}

	if all == true {
		err = c.builder.BuildAll(ctx, main, out)
		if err != nil {
			c.ui.Error(err.Error())
			return buildError
		}
	}

	return 0
}
