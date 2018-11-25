package config

import (
	"github.com/moorara/goto/config"
)

const (
	defaultLogLevel = "info"
)

// Config defines the configuration values
var Config = struct {
	LogLevel string
}{
	LogLevel: defaultLogLevel,
}

func init() {
	config.Pick(&Config)
}
