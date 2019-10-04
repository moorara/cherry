package command

import (
	"bytes"
	"context"
	"flag"
	"text/template"
	"time"

	"github.com/mitchellh/cli"
	"github.com/moorara/cherry/internal/action"
	"github.com/moorara/cherry/internal/spec"
	"github.com/moorara/cherry/pkg/cui"
)

const (
	buildErr     = 20
	buildFlagErr = 21

	buildTimeout = 2 * time.Minute

	buildSynopsis = `build artifacts`
	buildHelp     = `
	Use this command for building artifacts.
	Currently, this command can only build Go applications.

	Flags:

		-cross-compile:    build the binary for all platforms                (default: {{.Spec.Build.CrossCompile}})
		-main-file:        path to main.go file                              (default: {{.Spec.Build.MainFile}})
		-binary-file:      path for binary files                             (default: {{.Spec.Build.BinaryFile}})
		-version-package:  relative path to package containing version info  (default: {{.Spec.Build.VersionPackage}})

	Examples:

		cherry build
		cherry build -cross-compile
		cherry -main-file cmd/main.go -binary-file build/app
	`
)

// build is the build command.
type build struct {
	ui     cui.CUI
	fs     *flag.FlagSet
	action action.Action
}

// NewBuild creates a new build command.
func NewBuild(ui cui.CUI, workDir string, s spec.Spec) (cli.Command, error) {
	return &build{
		ui:     ui,
		fs:     s.Build.FlagSet(),
		action: action.NewBuild(ui, workDir, s),
	}, nil
}

// Synopsis returns a short one-line synopsis of the command.
func (c *build) Synopsis() string {
	return updateSynopsis
}

// Help returns a long help text including usage, description, and list of flags for the command.
func (c *build) Help() string {
	var buf bytes.Buffer
	t := template.Must(template.New("help").Parse(buildHelp))
	t.Execute(&buf, c)

	return buf.String()
}

// Run runs the actual command with the given command-line arguments.
func (c *build) Run(args []string) int {
	c.fs.Usage = func() {
		c.ui.Outputf(c.Help())
	}

	if err := c.fs.Parse(args); err != nil {
		return buildFlagErr
	}

	ctx, cancel := context.WithTimeout(context.Background(), buildTimeout)
	defer cancel()

	if err := c.action.Run(ctx); err != nil {
		c.ui.Errorf("%s", err)
		return buildErr
	}

	return 0
}
