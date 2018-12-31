package cli

import (
	"os"
	"path/filepath"

	"github.com/mitchellh/cli"
	"github.com/moorara/cherry/cmd/version"
	"github.com/moorara/cherry/internal/v1/formula"
	"github.com/moorara/cherry/internal/v1/spec"
)

// New creates a new command-line app
func New(ui cli.Ui, name, githubToken string) (*cli.CLI, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// If any error, we use a spec with default values
	spec, _ := spec.Read(spec.SpecFile)
	spec.SetDefaults()

	// Update spec
	name = filepath.Base(wd)
	spec.ToolVersion = version.Version
	spec.Build.BinaryFile = "bin/" + name

	formula, err := formula.New(ui, spec, wd, githubToken)
	if err != nil {
		return nil, err
	}

	app := cli.NewCLI(name, version.String())
	app.Args = os.Args[1:]
	app.Commands = map[string]cli.CommandFactory{
		"test": func() (cmd cli.Command, err error) {
			return NewTest(ui, spec, formula)
		},
		"build": func() (cmd cli.Command, err error) {
			return NewBuild(ui, spec, formula)
		},
		"release": func() (cmd cli.Command, err error) {
			return NewRelease(ui, spec, formula)
		},
	}

	return app, nil
}
