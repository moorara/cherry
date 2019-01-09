package cli

import (
	"os"

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
	s, err := spec.Read(spec.SpecFile)
	if err != nil {
		s = new(spec.Spec)
	}

	// Update spec
	s.SetDefaults()
	s.ToolVersion = version.Version

	f, err := formula.New(ui, s, wd, githubToken)
	if err != nil {
		return nil, err
	}

	app := cli.NewCLI(name, version.String())
	app.Args = os.Args[1:]
	app.Commands = map[string]cli.CommandFactory{
		"test": func() (cmd cli.Command, err error) {
			return NewTest(ui, s, f)
		},
		"build": func() (cmd cli.Command, err error) {
			return NewBuild(ui, s, f)
		},
		"release": func() (cmd cli.Command, err error) {
			return NewRelease(ui, s, f)
		},
	}

	return app, nil
}
