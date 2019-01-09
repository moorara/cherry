package cli

import (
	"bytes"
	"context"
	"text/template"
	"time"

	"github.com/mitchellh/cli"
	"github.com/moorara/cherry/internal/v1/formula"
	"github.com/moorara/cherry/internal/v1/spec"
)

type (
	// Build is the build CLI command
	Build struct {
		cli.Ui
		*spec.Spec
		formula.Formula
	}
)

const (
	buildError     = 20
	buildFlagError = 21
	buildTimeout   = 60 * time.Second
	buildSynopsis  = `build artifacts`
	buildHelp      = `
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

// NewBuild create a new build command
func NewBuild(ui cli.Ui, spec *spec.Spec, formula formula.Formula) (*Build, error) {
	cmd := &Build{
		Ui:      ui,
		Spec:    spec,
		Formula: formula,
	}

	return cmd, nil
}

// Synopsis returns the short one-line synopsis of the command.
func (c *Build) Synopsis() string {
	return buildSynopsis
}

// Help returns the long help text including usage, description, and list of flags for the command
func (c *Build) Help() string {
	var buf bytes.Buffer
	t := template.Must(template.New("help").Parse(buildHelp))
	t.Execute(&buf, c)

	return buf.String()
}

// Run runs the actual command with the given CLI instance and command-line arguments
func (c *Build) Run(args []string) int {
	// Parse command flags
	fs := c.Spec.Build.FlagSet()
	fs.Usage = func() { c.Ui.Output(c.Help()) }
	if err := fs.Parse(args); err != nil {
		return buildFlagError
	}

	ctx, cancel := context.WithTimeout(context.Background(), buildTimeout)
	defer cancel()

	err := c.Formula.Compile(ctx)
	if err != nil {
		c.Ui.Error(err.Error())
		return buildError
	}

	if c.Spec.Build.CrossCompile {
		_, err := c.Formula.CrossCompile(ctx)
		if err != nil {
			c.Ui.Error(err.Error())
			return buildError
		}
	}

	return 0
}
