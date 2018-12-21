package main

import (
	"os"

	"github.com/mitchellh/cli"
	"github.com/moorara/cherry/cmd/config"
	"github.com/moorara/cherry/cmd/version"
	"github.com/moorara/cherry/pkg/log"

	app "github.com/moorara/cherry/cmd/v1/cli"
	util "github.com/moorara/cherry/pkg/cli"
)

const (
	initError = 10
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
		ui = util.NewLoggerUI(logger)
	} else {
		ui = util.NewUI().Colored().Concurrent()
	}

	app, err := app.New(ui, config.Config.Name, version.String(), config.Config.GithubToken)
	if err != nil {
		ui.Error(err.Error())
		os.Exit(initError)
	}

	status, err := app.Run()
	if err != nil {
		ui.Error(err.Error())
	}

	os.Exit(status)
}
