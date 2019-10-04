package command

import (
	"bytes"
	"context"
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

		-cross-compile:    build the binary for all platforms                (default: {{.Build.CrossCompile}})
		-main-file:        path to main.go file                              (default: {{.Build.MainFile}})
		-binary-file:      path for binary files                             (default: {{.Build.BinaryFile}})
		-version-package:  relative path to package containing version info  (default: {{.Build.VersionPackage}})

	Examples:

		cherry build
		cherry build -cross-compile
		cherry -main-file cmd/main.go -binary-file build/app
	`
)

// build is the build command.
type build struct {
	ui     cui.CUI
	Build  spec.Build
	action action.Action
}

// NewBuild creates a new build command.
func NewBuild(ui cui.CUI, workDir string, s spec.Spec) (cli.Command, error) {
	return &build{
		ui:     ui,
		Build:  s.Build,
		action: action.NewBuild(ui, workDir, s),
	}, nil
}

// Synopsis returns a short one-line synopsis of the command.
func (c *build) Synopsis() string {
	return buildSynopsis
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
	fs := c.Build.FlagSet()
	fs.Usage = func() {
		c.ui.Outputf(c.Help())
	}

	if err := fs.Parse(args); err != nil {
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
