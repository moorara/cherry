package main

import (
	"os"

	"github.com/mitchellh/cli"
	"github.com/moorara/cherry/cmd/command"
	"github.com/moorara/cherry/cmd/config"
	"github.com/moorara/cherry/cmd/version"
	"github.com/moorara/cherry/pkg/log"
)

func main() {
	logger := log.NewJSONLogger(config.Config.Name, config.Config.LogLevel)
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

	// Uncomment this for the logger to be used safely by multiple goroutines
	// logger = logger.SyncLogger()

	ui := command.NewLoggerUI(logger)

	app := cli.NewCLI(config.Config.Name, version.Get())
	app.Args = os.Args[1:]
	app.Commands = map[string]cli.CommandFactory{
		"build": func() (cmd cli.Command, err error) {
			return command.NewBuild(ui), nil
		},
		"release": func() (cmd cli.Command, err error) {
			return command.NewRelease(ui), nil
		},
		"test": func() (cmd cli.Command, err error) {
			return command.NewTest(ui), nil
		},
	}

	status, err := app.Run()
	if err != nil {
		logger.Error("message", err.Error(), "error", err)
	}

	os.Exit(status)
}
