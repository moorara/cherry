package config

import (
	"github.com/moorara/goto/config"
)

const (
	defaultName     = "cherry"
	defaultLogLevel = "info"
)

// Config defines the configuration values
var Config = struct {
	Name     string
	LogLevel string
}{
	Name:     defaultName,
	LogLevel: defaultLogLevel,
}

func init() {
	config.Pick(&Config)
}
