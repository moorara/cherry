package main

import (
	"fmt"
	"os"

	"github.com/mitchellh/cli"
	"github.com/moorara/cherry/cmd/command"
	"github.com/moorara/cherry/cmd/config"
	"github.com/moorara/cherry/cmd/version"
	"github.com/moorara/cherry/pkg/log"
)

func main() {
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

	var ui cli.Ui
	if config.Config.LogJSON {
		ui = command.NewLoggerUI(logger)
	} else {
		ui = command.NewUI().Colored().Concurrent()
	}

	app := cli.NewCLI(config.Config.Name, version.String())
	app.Args = os.Args[1:]
	app.Commands = map[string]cli.CommandFactory{
		"build": func() (cmd cli.Command, err error) {
			return command.NewBuild(ui)
		},
		"release": func() (cmd cli.Command, err error) {
			return command.NewRelease(ui)
		},
		"test": func() (cmd cli.Command, err error) {
			return command.NewTest(ui)
		},
	}

	status, err := app.Run()
	if err != nil {
		ui.Error(fmt.Sprintf("An error occurred: %s", err))
	}

	os.Exit(status)
}
