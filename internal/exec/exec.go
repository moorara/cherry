package exec

import (
	"context"
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"
)

// Command executes a command and returns the stdout
func Command(ctx context.Context, wd string, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = wd

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", nil
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", nil
	}

	if err := cmd.Start(); err != nil {
		return "", err
	}

	outBytes, err := ioutil.ReadAll(stdout)
	if err != nil {
		return "", err
	}

	// Trim new line characters
	outStr := strings.Trim(string(outBytes), "\n")

	errBytes, err := ioutil.ReadAll(stderr)
	if err != nil {
		return "", err
	}

	// Trim new line characters
	errStr := strings.Trim(string(errBytes), "\n")

	err = cmd.Wait()
	if err != nil {
		return "", fmt.Errorf("%s: %s", err.Error(), errStr)
	}

	return outStr, nil
}
