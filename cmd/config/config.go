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
	LogLevel    string `flag:"-" env:"CHERRY_LOG_LEVEL" file:"CHERRY_LOG_LEVEL_FILE"`
	LogJSON     bool   `flag:"-" env:"CHERRY_LOG_JSON" file:"CHERRY_LOG_JSON_FILE"`
	GithubToken string `flag:"-" env:"CHERRY_GITHUB_TOKEN" file:"CHERRY_GITHUB_TOKEN_FILE"`
}{
	Name:     defaultName,
	LogLevel: defaultLogLevel,
	LogJSON:  defaultLogJSON,
}

func init() {
	config.Pick(&Config)
}
