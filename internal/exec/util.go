package exec

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
)

type (
	// ResetFunc is the type for a reset function
	ResetFunc func() error
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

// SetEnvVars set a set of environment variables and returns a reset function for restoring them
func SetEnvVars(keyVals ...string) (ResetFunc, error) {
	l := len(keyVals)
	if l%2 != 0 {
		return nil, errors.New("mismatching key-value pairs")
	}

	origs := make([]string, l)

	for i := 0; i < l; i += 2 {
		// Save original value of environment variable
		origs[i] = keyVals[i]
		origs[i+1] = os.Getenv(keyVals[i])

		// Set new value of environment variable
		if err := os.Setenv(keyVals[i], keyVals[i+1]); err != nil {
			return nil, err
		}
	}

	reset := func() error {
		for i := 0; i < len(origs); i += 2 {
			// Restore original value of environment variable
			if err := os.Setenv(origs[i], origs[i+1]); err != nil {
				return err
			}
		}
		return nil
	}

	return reset, nil
}
