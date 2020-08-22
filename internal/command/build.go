package command

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/mitchellh/cli"
	"github.com/moorara/cherry/internal/spec"
	"github.com/moorara/cherry/pkg/semver"
)

const (
	buildFlagErr = 201
	buildOSErr   = 202
	buildGitErr  = 203
	buildGoErr   = 204
	buildTimeout = 5 * time.Minute

	buildSynopsis = `build artifacts`
	buildHelp     = `
	Use this command for building artifacts.
	Currently, this command can only build Go applications.

	Flags:

		-cross-compile:    build the binary for all platforms                (default: {{.Spec.Build.CrossCompile}})
		-main-file:        path to main.go file                              (default: {{.Spec.Build.MainFile}})
		-binary-file:      path for binary files                             (default: {{.Spec.Build.BinaryFile}})
		-version-package:  relative path to package containing version info  (default: {{.Spec.Build.VersionPackage}})

	Examples:

		cherry build
		cherry build -cross-compile
		cherry -main-file cmd/my-app/main.go -binary-file build/my-app
	`
)

// buildCommand implements cli.Command interface.
type buildCommand struct {
	ui        cli.Ui
	spec      spec.Spec
	artifacts []string
}

// NewBuildCommand creates a build command.
func NewBuildCommand(ui cli.Ui, s spec.Spec) (cli.Command, error) {
	return &buildCommand{
		ui:        ui,
		spec:      s,
		artifacts: []string{},
	}, nil
}

// Synopsis returns a short one-line synopsis of the command.
func (c *buildCommand) Synopsis() string {
	return buildSynopsis
}

// Help returns a long help text including usage, description, and list of flags for the command.
func (c *buildCommand) Help() string {
	var buf bytes.Buffer
	t := template.Must(template.New("help").Parse(buildHelp))
	_ = t.Execute(&buf, c)
	return buf.String()
}

// Run runs the actual command with the given command-line arguments.
func (c *buildCommand) Run(args []string) int {
	fs := c.spec.Build.FlagSet()
	fs.Usage = func() {
		c.ui.Output(c.Help())
	}

	if err := fs.Parse(args); err != nil {
		return buildFlagErr
	}

	ctx, cancel := context.WithTimeout(context.Background(), buildTimeout)
	defer cancel()

	// Run preflight checks

	var dir string

	{
		// c.ui.Output("◉ Running preflight checks ...")

		var err error
		dir, err = os.Getwd()
		if err != nil {
			c.ui.Error(fmt.Sprintf("Error on getting the current working directory: %s", err))
			return buildOSErr
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
			return buildGitErr
		}

		stdout.Reset()
		stderr.Reset()
		cmd = exec.CommandContext(ctx, "go", "version")
		cmd.Dir = dir
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			c.ui.Error(fmt.Sprintf("Error on checking go: %s %s", err, strings.Trim(stderr.String(), "\n")))
			return buildGoErr
		}
	}

	// Get git information

	var gitStatusClean bool
	var gitSHA, gitShortSHA, gitBranch string

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
		cmd = exec.CommandContext(ctx, "git", "rev-parse", "HEAD")
		cmd.Dir = dir
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			c.ui.Error(fmt.Sprintf("Error on running git rev-parse HEAD: %s %s", err, strings.Trim(stderr.String(), "\n")))
			return buildGitErr
		}
		gitSHA = strings.Trim(stdout.String(), "\n")
		gitShortSHA = gitSHA[:7]

		stdout.Reset()
		stderr.Reset()
		cmd = exec.CommandContext(ctx, "git", "rev-parse", "--abbrev-ref", "HEAD")
		cmd.Dir = dir
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			c.ui.Error(fmt.Sprintf("Error on running git rev-parse --abbrev-ref HEAD: %s %s", err, strings.Trim(stderr.String(), "\n")))
			return buildGitErr
		}
		gitBranch = strings.Trim(stdout.String(), "\n")
	}

	// Resolve the current semantic version

	var version semver.SemVer

	{
		releaseRE := regexp.MustCompile(`^v?([0-9]+)\.([0-9]+)\.([0-9]+)$`)
		prereleaseRE := regexp.MustCompile(`^(v?([0-9]+)\.([0-9]+)\.([0-9]+))-([0-9]+)-g([0-9a-f]+)$`)

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

		if len(gitDescribe) == 0 {
			// No git tag and no previous semantic version -> using the default initial semantic version
			version = semver.SemVer{
				Major:      0,
				Minor:      1,
				Patch:      0,
				Prerelease: []string{gitShortSHA},
			}
		} else if subs := releaseRE.FindStringSubmatch(gitDescribe); len(subs) == 4 {
			// The tag points to the HEAD commit
			// Example: v0.2.7 --> subs = []string{"v0.2.7", "0", "2", "7"}
			version, _ = semver.Parse(subs[0])
		} else if subs := prereleaseRE.FindStringSubmatch(gitDescribe); len(subs) == 7 {
			// The tag is the most recent tag reachable from the HEAD commit
			// Example: v0.2.7-10-gabcdeff --> subs = []string{"v0.2.7-10-gabcdeff", "v0.2.7", "0", "2", "7", "10", "abcdeff"}
			version, _ = semver.Parse(subs[1])
			version.AddPrerelease(subs[5], subs[6])
		}

		if !gitStatusClean {
			version.AddPrerelease("dev")
		}
	}

	// Get go compiler information

	var goVersion string

	{
		var stdout, stderr bytes.Buffer
		cmd := exec.CommandContext(ctx, "go", "version")
		cmd.Dir = dir
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			c.ui.Error(fmt.Sprintf("Error on running go version: %s %s", err, strings.Trim(stderr.String(), "\n")))
			return buildGoErr
		}
		goVersion = regexp.MustCompile(`go\d+\.\d+(\.\d+)?`).FindString(stdout.String())
	}

	// Resolve the full import path to the version package

	var versionPkg string

	{
		var stdout, stderr bytes.Buffer
		cmd := exec.CommandContext(ctx, "go", "list", c.spec.Build.VersionPackage)
		cmd.Dir = dir
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			c.ui.Error(fmt.Sprintf("Error on running go list: %s %s", err, strings.Trim(stderr.String(), "\n")))
			return buildGoErr
		}
		versionPkg = strings.Trim(stdout.String(), "\n")
	}

	// Construct LD flags

	var ldFlags string

	{
		buildTool := fmt.Sprintf("%s %s", c.spec.ToolName, c.spec.ToolVersion)
		buildTime := time.Now().UTC().Format(time.RFC3339Nano)

		versionFlag := fmt.Sprintf("-X %s.Version=%s", versionPkg, version)
		commitFlag := fmt.Sprintf("-X %s.Commit=%s", versionPkg, gitShortSHA)
		branchFlag := fmt.Sprintf("-X %s.Branch=%s", versionPkg, gitBranch)
		goVersionFlag := fmt.Sprintf("-X %s.GoVersion=%s", versionPkg, goVersion)
		buildToolFlag := fmt.Sprintf("-X %s.BuildTool=%s", versionPkg, buildTool)
		buildTimeFlag := fmt.Sprintf("-X %s.BuildTime=%s", versionPkg, buildTime)
		ldFlags = fmt.Sprintf("%s %s %s %s %s %s", versionFlag, commitFlag, branchFlag, goVersionFlag, buildToolFlag, buildTimeFlag)
	}

	// Build binaries
	{
		if !c.spec.Build.CrossCompile {
			err := c.build(ctx, dir, ldFlags, c.spec.Build.BinaryFile)
			if err != nil {
				c.ui.Error(fmt.Sprintf("Error on building binary: %s", err))
				return buildGoErr
			}

			c.artifacts = append(c.artifacts, c.spec.Build.BinaryFile)
		} else {
			// Cross-compiling
			for _, platform := range c.spec.Build.Platforms {
				vals := strings.Split(platform, "-")
				if err := os.Setenv("GOOS", vals[0]); err != nil {
					c.ui.Error(fmt.Sprintf("Error on setting environment variable GOOS: %s", err))
					return buildOSErr
				}
				if err := os.Setenv("GOARCH", vals[1]); err != nil {
					c.ui.Error(fmt.Sprintf("Error on setting environment variable GOARCH: %s", err))
					return buildOSErr
				}

				binFile := fmt.Sprintf("%s-%s", c.spec.Build.BinaryFile, platform)
				err := c.build(ctx, dir, ldFlags, binFile)
				if err != nil {
					c.ui.Error(fmt.Sprintf("Error on building binary: %s", err))
					return buildGoErr
				}

				c.artifacts = append(c.artifacts, binFile)
			}

			os.Unsetenv("GOOS")
			os.Unsetenv("GOARCH")
		}
	}

	return 0
}

func (c *buildCommand) build(ctx context.Context, dir, ldFlags, binFile string) error {
	args := []string{"build"}
	if ldFlags != "" {
		args = append(args, "-ldflags", ldFlags)
	}
	if binFile != "" {
		args = append(args, "-o", binFile)
	}
	args = append(args, c.spec.Build.MainFile)

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = dir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s %s", err, strings.Trim(stderr.String(), "\n"))
	}

	c.ui.Info(fmt.Sprintf("🍒 %s", binFile))

	return nil
}
