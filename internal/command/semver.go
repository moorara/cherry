package command

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/mitchellh/cli"
	"github.com/moorara/cherry/pkg/semver"
)

const (
	semverFlagErr = 201
	semverOSErr   = 202
	semverGitErr  = 203
	semverTimeout = 10 * time.Second

	semverSynopsis = `get semantic version`
	semverHelp     = `
	Use this command for getting the current semantic version.

	Examples:

		cherry semver
	`
)

// semverCommand implements cli.Command interface.
type semverCommand struct {
	ui      cli.Ui
	version semver.SemVer
}

// NewSemverCommand creates a semver command.
func NewSemverCommand(ui cli.Ui) (cli.Command, error) {
	return &semverCommand{
		ui: ui,
	}, nil
}

// Synopsis returns a short one-line synopsis of the command.
func (c *semverCommand) Synopsis() string {
	return semverSynopsis
}

// Help returns a long help text including usage, description, and list of flags for the command.
func (c *semverCommand) Help() string {
	return semverHelp
}

// Run runs the actual command with the given command-line arguments.
func (c *semverCommand) Run(args []string) int {
	fs := flag.NewFlagSet("semver", flag.ContinueOnError)
	fs.Usage = func() {
		c.ui.Output(c.Help())
	}

	if err := fs.Parse(args); err != nil {
		return semverFlagErr
	}

	ctx, cancel := context.WithTimeout(context.Background(), semverTimeout)
	defer cancel()

	// Run preflight checks

	var dir string

	{
		// c.ui.Output("◉ Running preflight checks ...")

		var err error
		dir, err = os.Getwd()
		if err != nil {
			c.ui.Error(fmt.Sprintf("Error on getting the current working directory: %s", err))
			return semverOSErr
		}
	}

	{
		var stdout, stderr bytes.Buffer
		cmd := exec.CommandContext(ctx, "git", "version")
		cmd.Dir = dir
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			c.ui.Error(fmt.Sprintf("Error on checking git: %s %s", err, strings.Trim(stderr.String(), "\n")))
			return semverGitErr
		}
	}

	// Get git information

	var gitStatusClean bool
	var gitCommitCount string
	var gitSHA, gitShortSHA string

	{
		var stdout, stderr bytes.Buffer
		cmd := exec.CommandContext(ctx, "git", "status", "--porcelain")
		cmd.Dir = dir
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			c.ui.Error(fmt.Sprintf("Error on running git status --porcelain: %s %s", err, strings.Trim(stderr.String(), "\n")))
			return semverGitErr
		}
		gitStatusClean = len(stdout.String()) == 0

		stdout.Reset()
		stderr.Reset()
		cmd = exec.CommandContext(ctx, "git", "rev-list", "--count", "HEAD")
		cmd.Dir = dir
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			c.ui.Error(fmt.Sprintf("Error on running git rev-list --count HEAD: %s %s", err, strings.Trim(stderr.String(), "\n")))
			return semverGitErr
		}
		gitCommitCount = strings.Trim(stdout.String(), "\n")

		stdout.Reset()
		stderr.Reset()
		cmd = exec.CommandContext(ctx, "git", "rev-parse", "HEAD")
		cmd.Dir = dir
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			c.ui.Error(fmt.Sprintf("Error on running git rev-parse HEAD: %s %s", err, strings.Trim(stderr.String(), "\n")))
			return semverGitErr
		}
		gitSHA = strings.Trim(stdout.String(), "\n")
		gitShortSHA = gitSHA[:7]
	}

	// Resolve the current semantic version
	{
		var stdout, stderr bytes.Buffer
		cmd := exec.CommandContext(ctx, "git", "describe", "--tags", "HEAD")
		cmd.Dir = dir
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			// 128 is returned when there is no git tag
			if exiterr, ok := err.(*exec.ExitError); !ok || exiterr.ExitCode() != 128 {
				c.ui.Error(fmt.Sprintf("Error on running git describe --tags HEAD: %s %s", err, strings.Trim(stderr.String(), "\n")))
				return semverGitErr
			}
		}
		gitDescribe := strings.Trim(stdout.String(), "\n")

		releaseRE := regexp.MustCompile(`^v?([0-9]+)\.([0-9]+)\.([0-9]+)$`)
		prereleaseRE := regexp.MustCompile(`^(v?([0-9]+)\.([0-9]+)\.([0-9]+))-([0-9]+)-g([0-9a-f]+)$`)

		if len(gitDescribe) == 0 {
			// No git tag and no previous semantic version -> using the default initial semantic version

			c.version = semver.SemVer{
				Major: 0, Minor: 1, Patch: 0,
				Prerelease: []string{gitCommitCount},
			}

			if gitStatusClean {
				c.version.AddPrerelease(gitShortSHA)
			} else {
				c.version.AddPrerelease("dev")
			}
		} else if subs := releaseRE.FindStringSubmatch(gitDescribe); len(subs) == 4 {
			// The tag points to the HEAD commit
			// Example: v0.2.7 --> subs = []string{"v0.2.7", "0", "2", "7"}

			c.version, _ = semver.Parse(subs[0])

			if !gitStatusClean {
				c.version = c.version.Next()
				c.version.AddPrerelease("0", "dev")
			}
		} else if subs := prereleaseRE.FindStringSubmatch(gitDescribe); len(subs) == 7 {
			// The tag is the most recent tag reachable from the HEAD commit
			// Example: v0.2.7-10-gabcdeff --> subs = []string{"v0.2.7-10-gabcdeff", "v0.2.7", "0", "2", "7", "10", "abcdeff"}

			c.version, _ = semver.Parse(subs[1])
			c.version = c.version.Next()
			c.version.AddPrerelease(subs[5])

			if gitStatusClean {
				c.version.AddPrerelease(subs[6])
			} else {
				c.version.AddPrerelease("dev")
			}
		}

		c.ui.Output(c.version.String())
	}

	return 0
}
