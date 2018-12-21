package cli

import (
	"os"
	"path/filepath"

	"github.com/mitchellh/cli"
	"github.com/moorara/cherry/internal/v1/spec"
)

// New creates a new command-line app
func New(ui cli.Ui, name, version, githubToken string) (*cli.CLI, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// If any error, we use a spec with default values
	spec, _ := spec.Read(spec.SpecFile)
	spec.SetDefaults()

	// Update spec
	name = filepath.Base(wd)
	spec.Build.BinaryFile = "bin/" + name

	app := cli.NewCLI(name, version)
	app.Args = os.Args[1:]
	app.Commands = map[string]cli.CommandFactory{
		"test": func() (cmd cli.Command, err error) {
			return NewTest(ui, spec, wd)
		},
		"build": func() (cmd cli.Command, err error) {
			return NewBuild(ui, spec, wd)
		},
		"release": func() (cmd cli.Command, err error) {
			return NewRelease(ui, spec, wd, githubToken)
		},
	}

	return app, nil
}
