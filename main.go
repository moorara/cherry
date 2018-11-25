package main

import (
	"github.com/moorara/cherry/cmd/config"
	"github.com/moorara/cherry/cmd/version"
	"github.com/moorara/cherry/pkg/log"
)

func main() {
	logger := log.NewLogger("cherry", config.Config.LogLevel)

	logger.Info(
		"version", version.Version,
		"revision", version.Revision,
		"branch", version.Branch,
		"goVersion", version.GoVersion,
		"buildTime", version.BuildTime,
		"buildTool", version.BuildTool,
		"message", "Cherry started.",
	)
}
