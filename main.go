package main

import (
	"os"

	"github.com/mitchellh/cli"
	"github.com/moorara/cherry/cmd/command"
	"github.com/moorara/cherry/cmd/config"
	"github.com/moorara/cherry/cmd/version"
	"github.com/moorara/cherry/pkg/log"
)

const (
	workDirError = 10
)

func main() {
	// Create logger
	logger := log.NewJSONLogger(config.Config.Name, config.Config.LogLevel)
	logger = logger.SyncLogger()
	logger = logger.With(
		config.Config.Name, map[string]string{
			"version":   version.Version,
			"revision":  version.Revision,
			"branch":    version.Branch,
			"goVersion": version.GoVersion,
			"buildTool": version.BuildTool,
			"buildTime": version.BuildTime,
		},
	)

	// Create cli.Ui
	var ui cli.Ui
	if config.Config.LogJSON {
		ui = command.NewLoggerUI(logger)
	} else {
		ui = command.NewUI().Colored().Concurrent()
	}

	// Get working directory
	wd, err := os.Getwd()
	if err != nil {
		ui.Error(err.Error())
		os.Exit(workDirError)
	}

	// Create cli app
	app := cli.NewCLI(config.Config.Name, version.String())
	app.Args = os.Args[1:]
	app.Commands = map[string]cli.CommandFactory{
		"test": func() (cmd cli.Command, err error) {
			return command.NewTest(ui, wd)
		},
		"build": func() (cmd cli.Command, err error) {
			return command.NewBuild(ui, wd)
		},
		"release": func() (cmd cli.Command, err error) {
			return command.NewRelease(ui, wd, config.Config.GithubToken)
		},
	}

	status, err := app.Run()
	if err != nil {
		ui.Error(err.Error())
	}

	os.Exit(status)
}
