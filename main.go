package main

import (
	"errors"
	"os"

	"github.com/mitchellh/cli"
	"github.com/moorara/cherry/cmd/command"
	"github.com/moorara/cherry/cmd/version"
	"github.com/moorara/cherry/internal/spec"
	"github.com/moorara/cherry/pkg/cui"
	"github.com/moorara/konfig"
)

const (
	osErr     = 10
	configErr = 11
	specErr   = 12
)

var config = struct {
	GithubToken string
}{}

func main() {
	ui := cui.New()

	wd, err := os.Getwd()
	if err != nil {
		ui.Errorf("%s", err)
		os.Exit(osErr)
	}

	err = konfig.Pick(&config, konfig.PrefixEnv("CHERRY_"), konfig.PrefixFileEnv("CHERRY_"))
	if err != nil {
		ui.Errorf("%s", err)
		os.Exit(configErr)
	}

	// Read the spec
	// If spec file not found, create a default spec
	s, err := spec.Read()
	if err != nil {
		var se *spec.Error
		if errors.As(err, &se) && se.SpecNotFound {
			s = new(spec.Spec)
		} else {
			ui.Errorf("%s", err)
			os.Exit(specErr)
		}
	}

	// Update spec
	s.SetDefaults()
	s.ToolVersion = version.Version

	c := cli.NewCLI("cherry", version.String())
	c.Args = os.Args[1:]
	c.Commands = map[string]cli.CommandFactory{
		"build": func() (cli.Command, error) {
			return command.NewBuild(ui, wd, *s)
		},
		"update": func() (cli.Command, error) {
			return command.NewUpdate(ui, config.GithubToken)
		},
	}

	code, err := c.Run()
	if err != nil {
		ui.Errorf("%s", err)
	}

	os.Exit(code)
}
