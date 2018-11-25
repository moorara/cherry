package log

import (
	"os"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

// Logger wraps a go-kit Logger
type Logger struct {
	Name   string
	Logger log.Logger
}

// NewLogger creates a new logger
func NewLogger(logger log.Logger, name, logLevel string) *Logger {
	logger = log.With(
		logger,
		"logger", name,
		"timestamp", log.DefaultTimestampUTC,
	)

	logLevel = strings.ToLower(logLevel)
	switch logLevel {
	case "debug":
		logger = level.NewFilter(logger, level.AllowDebug())
	case "info":
		logger = level.NewFilter(logger, level.AllowInfo())
	case "warn":
		logger = level.NewFilter(logger, level.AllowWarn())
	case "error":
		logger = level.NewFilter(logger, level.AllowError())
	case "none":
		logger = level.NewFilter(logger, level.AllowNone())
	}

	return &Logger{
		Name:   name,
		Logger: logger,
	}
}

// NewNopLogger creates a new logger for testing purposes
func NewNopLogger() *Logger {
	logger := log.NewNopLogger()
	return NewLogger(logger, "nop", "none")
}

// NewJSONLogger creates a new logger logging in JSON
func NewJSONLogger(name, logLevel string) *Logger {
	logger := log.NewJSONLogger(os.Stdout)
	return NewLogger(logger, name, logLevel)
}

// NewFmtLogger creates a new logger logging using fmt format strings
func NewFmtLogger(name, logLevel string) *Logger {
	logger := log.NewLogfmtLogger(os.Stdout)
	return NewLogger(logger, name, logLevel)
}

// With returns a new logger which always logs a set of key-value pairs
func (l *Logger) With(kv ...interface{}) *Logger {
	return &Logger{
		Name:   l.Name,
		Logger: log.With(l.Logger, kv...),
	}
}

// SyncLogger returns a new logger which can be used concurrently by goroutines.
// Only one goroutine is allowed to log at a time and other goroutines will block until the logger is available.
func (l *Logger) SyncLogger() *Logger {
	return &Logger{
		Name:   l.Name,
		Logger: log.NewSyncLogger(l.Logger),
	}
}

// Debug logs in debug level
func (l *Logger) Debug(kv ...interface{}) error {
	return level.Debug(l.Logger).Log(kv...)
}

// Info logs in info level
func (l *Logger) Info(kv ...interface{}) error {
	return level.Info(l.Logger).Log(kv...)
}

// Warn logs in warn level
func (l *Logger) Warn(kv ...interface{}) error {
	return level.Warn(l.Logger).Log(kv...)
}

// Error logs in error level
func (l *Logger) Error(kv ...interface{}) error {
	return level.Error(l.Logger).Log(kv...)
}
