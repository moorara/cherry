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
	buildFlagErr = 21
	buildOSErr   = 22
	buildGoErr   = 23
	buildGitErr  = 24
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

// build implements cli.Command interface.
type build struct {
	ui   cli.Ui
	spec spec.Spec
}

// NewBuild creates a build command.
func NewBuild(ui cli.Ui, s spec.Spec) (cli.Command, error) {
	return &build{
		ui:   ui,
		spec: s,
	}, nil
}

// Synopsis returns a short one-line synopsis of the command.
func (b *build) Synopsis() string {
	return buildSynopsis
}

// Help returns a long help text including usage, description, and list of flags for the command.
func (b *build) Help() string {
	var buf bytes.Buffer
	t := template.Must(template.New("help").Parse(buildHelp))
	_ = t.Execute(&buf, b)
	return buf.String()
}

// Run runs the actual command with the given command-line arguments.
func (b *build) Run(args []string) int {
	fs := b.spec.Build.FlagSet()
	fs.Usage = func() {
		b.ui.Output(b.Help())
	}

	if err := fs.Parse(args); err != nil {
		return buildFlagErr
	}

	ctx, cancel := context.WithTimeout(context.Background(), updateTimeout)
	defer cancel()

	dir, err := os.Getwd()
	if err != nil {
		b.ui.Error(fmt.Sprintf("Error on getting the current working directory: %s", err))
		return buildOSErr
	}

	// Get git information

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "git", "status", "--porcelain")
	cmd.Dir = dir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		b.ui.Error(fmt.Sprintf("Error on running git status --porcelain: %s %s", err.Error(), strings.Trim(stderr.String(), "\n")))
		return buildGitErr
	}
	gitStatusClean := len(stdout.String()) == 0

	stdout.Reset()
	stderr.Reset()
	cmd = exec.CommandContext(ctx, "git", "rev-parse", "HEAD")
	cmd.Dir = dir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		b.ui.Error(fmt.Sprintf("Error on running git rev-parse HEAD: %s %s", err.Error(), strings.Trim(stderr.String(), "\n")))
		return buildGitErr
	}
	gitSHA := strings.Trim(stdout.String(), "\n")
	gitShortSHA := gitSHA[:7]

	stdout.Reset()
	stderr.Reset()
	cmd = exec.CommandContext(ctx, "git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = dir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		b.ui.Error(fmt.Sprintf("Error on running git rev-parse --abbrev-ref HEAD: %s %s", err.Error(), strings.Trim(stderr.String(), "\n")))
		return buildGitErr
	}
	gitBranch := strings.Trim(stdout.String(), "\n")

	// Resolve the current semantic version

	var version semver.SemVer
	releaseRE := regexp.MustCompile(`^v?([0-9]+)\.([0-9]+)\.([0-9]+)$`)
	prereleaseRE := regexp.MustCompile(`^(v?([0-9]+)\.([0-9]+)\.([0-9]+))-([0-9]+)-g([0-9a-f]+)$`)

	stdout.Reset()
	stderr.Reset()
	cmd = exec.CommandContext(ctx, "git", "describe", "--tags", "HEAD")
	cmd.Dir = dir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		b.ui.Error(fmt.Sprintf("Error on running git describe --tags HEAD: %s %s", err.Error(), strings.Trim(stderr.String(), "\n")))
		return buildGitErr
	}
	gitDescribe := strings.Trim(stdout.String(), "\n")

	if len(gitDescribe) == 0 {
		// No git tag and no previous semantic version -> use the default initial semantic version
		version = semver.SemVer{
			Major:      0,
			Minor:      1,
			Patch:      0,
			Prerelease: []string{gitShortSHA},
		}
	} else if subs := releaseRE.FindStringSubmatch(gitDescribe); len(subs) == 4 {
		// The tag points to the HEAD commit
		// example: v0.2.7 --> subs = []string{"v0.2.7", "0", "2", "7"}
		version, _ = semver.Parse(subs[0])
	} else if subs := prereleaseRE.FindStringSubmatch(gitDescribe); len(subs) == 7 {
		// The tag is the most recent tag reachable from the HEAD commit
		// example: v0.2.7-10-gabcdeff --> subs = []string{"v0.2.7-10-gabcdeff", "v0.2.7", "0", "2", "7", "10", "abcdeff"}
		version, _ = semver.Parse(subs[1])
		version.AddPrerelease(subs[5], subs[6])
	}

	if !gitStatusClean {
		version.AddPrerelease("dev")
	}

	// Get go compiler information

	stdout.Reset()
	stderr.Reset()
	cmd = exec.CommandContext(ctx, "go", "version")
	cmd.Dir = dir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		b.ui.Error(fmt.Sprintf("Error on running go version: %s %s", err.Error(), strings.Trim(stderr.String(), "\n")))
		return buildGoErr
	}
	goVersion := regexp.MustCompile(`go\d+\.\d+(\.\d+)?`).FindString(stdout.String())

	// Resolve the full import path to the version package

	stdout.Reset()
	stderr.Reset()
	cmd = exec.CommandContext(ctx, "go", "list", b.spec.Build.VersionPackage)
	cmd.Dir = dir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		b.ui.Error(fmt.Sprintf("Error on running go list: %s %s", err.Error(), strings.Trim(stderr.String(), "\n")))
		return buildGoErr
	}
	versionPkg := strings.Trim(stdout.String(), "\n")

	// Construct LD flags

	buildTool := fmt.Sprintf("%s %s", b.spec.ToolName, b.spec.ToolVersion)
	buildTime := time.Now().UTC().Format(time.RFC3339Nano)

	versionFlag := fmt.Sprintf("-X %s.Version=%s", versionPkg, version)
	revisionFlag := fmt.Sprintf("-X %s.Revision=%s", versionPkg, gitShortSHA)
	branchFlag := fmt.Sprintf("-X %s.Branch=%s", versionPkg, gitBranch)
	goVersionFlag := fmt.Sprintf("-X %s.GoVersion=%s", versionPkg, goVersion)
	buildToolFlag := fmt.Sprintf("-X %s.BuildTool=%s", versionPkg, buildTool)
	buildTimeFlag := fmt.Sprintf("-X %s.BuildTime=%s", versionPkg, buildTime)
	ldFlags := fmt.Sprintf("%s %s %s %s %s %s", versionFlag, revisionFlag, branchFlag, goVersionFlag, buildToolFlag, buildTimeFlag)

	// Build single binary

	if len(b.spec.Build.Platforms) == 0 {
		err := b.build(ctx, dir, ldFlags, b.spec.Build.BinaryFile)
		if err != nil {
			b.ui.Error(fmt.Sprintf("Error on building binary: %s", err.Error()))
			return buildGoErr
		}
		return 0
	}

	// Cross-compile binaries

	for _, platform := range b.spec.Build.Platforms {
		vals := strings.Split(platform, "-")
		if err := os.Setenv("GOOS", vals[0]); err != nil {
			b.ui.Error(fmt.Sprintf("Error on setting environment variable GOOS: %s", err.Error()))
			return buildOSErr
		}
		if err := os.Setenv("GOARCH", vals[1]); err != nil {
			b.ui.Error(fmt.Sprintf("Error on setting environment variable GOARCH: %s", err.Error()))
			return buildOSErr
		}

		binFile := fmt.Sprintf("%s-%s", b.spec.Build.BinaryFile, platform)
		err := b.build(ctx, dir, ldFlags, binFile)
		if err != nil {
			b.ui.Error(fmt.Sprintf("Error on building binary: %s", err.Error()))
			return buildGoErr
		}
	}

	os.Unsetenv("GOOS")
	os.Unsetenv("GOARCH")

	return 0
}

func (b *build) build(ctx context.Context, dir, ldFlags, binFile string) error {
	args := []string{"build"}
	if ldFlags != "" {
		args = append(args, "-ldflags", ldFlags)
	}
	if binFile != "" {
		args = append(args, "-o", binFile)
	}
	args = append(args, b.spec.Build.MainFile)

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = dir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s %s", err.Error(), strings.Trim(stderr.String(), "\n"))
	}

	b.ui.Info(fmt.Sprintf("🍒 %s", binFile))

	return nil
}
