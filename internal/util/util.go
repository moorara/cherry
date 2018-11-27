package util

import (
	"fmt"
	"os"
	"os/exec"
)

// EnsureCommands ensures the given commands are available
func EnsureCommands(names ...string) error {
	for _, name := range names {
		err := exec.Command("which", name).Run()
		if err != nil {
			return fmt.Errorf("%s command is no available", name)
		}
	}

	return nil
}

// EnsureEnvVars ensures the given environment variables are set
func EnsureEnvVars(names ...string) error {
	for _, name := range names {
		val := os.Getenv(name)
		if val == "" {
			return fmt.Errorf("%s environment variable is not set", name)
		}
	}

	return nil
}
