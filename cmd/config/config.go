package config

import (
	"github.com/moorara/goto/config"
)

const (
	defaultName     = "cherry"
	defaultLogLevel = "info"
	defaultLogJSON  = false
)

// Config defines the configuration values
var Config = struct {
	Name        string `flag:"-" env:"-" file:"-"`
	LogLevel    string `flag:"log.level" env:"CHERRY_LOG_LEVEL" file:"CHERRY_LOG_LEVEL_FILE"`
	LogJSON     bool   `flag:"log.json" env:"CHERRY_LOG_JSON" file:"CHERRY_LOG_JSON_FILE"`
	GithubToken string `flag:"github.token" env:"CHERRY_GITHUB_TOKEN" file:"CHERRY_GITHUB_TOKEN_FILE"`
}{
	Name:     defaultName,
	LogLevel: defaultLogLevel,
	LogJSON:  defaultLogJSON,
}

func init() {
	config.Pick(&Config)
}
