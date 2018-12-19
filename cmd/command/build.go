package command

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/mitchellh/cli"
	"github.com/moorara/cherry/internal/git"
	"github.com/moorara/cherry/internal/util"
)

type (
	buildInfo struct {
		Version   string
		Revision  string
		Branch    string
		GoVersion string
		BuildTool string
		BuildTime time.Time
	}

	// Build is the build CLI command
	Build struct {
		ui      cli.Ui
		workDir string
		git     git.Git
	}
)

const (
	buildTool = "Cherry"

	buildError     = 20
	buildFlagError = 21
	buildTimeout   = 1 * time.Minute
	buildSynopsis  = `Build artifacts`
	buildHelp      = `
	Use this command for building artifacts.

	Flags:
		-main: path to main.go file                (default: main.go)
		-out:  path for binary files               (default: bin/app)
		-all:  build the binary for all platforms  (default: false)
	`
)

var (
	platforms = []string{"linux-386", "linux-amd64", "darwin-386", "darwin-amd64", "windows-386", "windows-amd64"}
)

// NewBuild create a new build command
func NewBuild(ui cli.Ui, workDir string) (*Build, error) {
	cmd := &Build{
		ui:      ui,
		workDir: workDir,
		git:     git.New(workDir),
	}

	return cmd, nil
}

func (c *Build) getBuildInfo(ctx context.Context) (*buildInfo, error) {
	info := new(buildInfo)

	version, err := ioutil.ReadFile(filepath.Join(c.workDir, versionFile))
	if err != nil {
		return nil, err
	}

	info.Version = strings.Trim(string(version), "\n")

	info.Revision, err = c.git.GetCommitSHA(true)
	if err != nil {
		return nil, err
	}

	info.Branch, err = c.git.GetBranchName()
	if err != nil {
		return nil, err
	}

	info.GoVersion = runtime.Version()
	info.BuildTool = buildTool
	info.BuildTime = time.Now().UTC()

	return info, nil
}

func (c *Build) getLDFlags(ctx context.Context, info *buildInfo) (string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := exec.CommandContext(ctx, "go", "list", versionPackage)
	cmd.Dir = c.workDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%s: %s", err.Error(), stderr.String())
	}

	vPkg := strings.Trim(stdout.String(), "\n")

	versionFlag := fmt.Sprintf("-X %s.Version=%s", vPkg, info.Version)
	revisionFlag := fmt.Sprintf("-X %s.Revision=%s", vPkg, info.Revision)
	branchFlag := fmt.Sprintf("-X %s.Branch=%s", vPkg, info.Branch)
	goVersionFlag := fmt.Sprintf("-X %s.GoVersion=%s", vPkg, info.GoVersion)
	buildToolFlag := fmt.Sprintf("-X %s.BuildTool=%s", vPkg, info.BuildTool)
	buildTimeFlag := fmt.Sprintf("-X %s.BuildTime=%s", vPkg, info.BuildTime.Format(time.RFC3339Nano))

	ldflags := fmt.Sprintf("%s %s %s %s %s %s", versionFlag, revisionFlag, branchFlag, goVersionFlag, buildToolFlag, buildTimeFlag)

	return ldflags, nil
}

func (c *Build) build(ctx context.Context, main, out string) error {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	info, err := c.getBuildInfo(ctx)
	if err != nil {
		return err
	}

	ldflags, err := c.getLDFlags(ctx, info)
	if err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, "go", "build", "-ldflags", ldflags, "-o", out, main)
	cmd.Dir = c.workDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %s", err.Error(), stderr.String())
	}

	c.ui.Info(out)

	return nil
}

func (c *Build) buildAll(ctx context.Context, main, outPrefix string) error {
	info, err := c.getBuildInfo(ctx)
	if err != nil {
		return err
	}

	ldflags, err := c.getLDFlags(ctx, info)
	if err != nil {
		return err
	}

	for _, platform := range platforms {
		var stdout bytes.Buffer
		var stderr bytes.Buffer

		env := strings.Split(platform, "-")
		reset, err := util.SetEnvVars("GOOS", env[0], "GOARCH", env[1])
		if err != nil {
			return err
		}

		stdout.Reset()
		stderr.Reset()

		out := fmt.Sprintf("%s-%s", outPrefix, platform)
		cmd := exec.CommandContext(ctx, "go", "build", "-ldflags", ldflags, "-o", out, main)
		cmd.Dir = c.workDir
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("%s: %s", err.Error(), stderr.String())
		}

		// Restore environment variables
		reset()

		c.ui.Info(out)
	}

	return nil
}

// Synopsis returns the short one-line synopsis of the command.
func (c *Build) Synopsis() string {
	return buildSynopsis
}

// Help returns the long help text including usage, description, and list of flags for the command
func (c *Build) Help() string {
	return buildHelp
}

// Run runs the actual command with the given CLI instance and command-line arguments
func (c *Build) Run(args []string) int {
	var main, out string
	var all bool

	_, repoName, err := c.git.GetRepoName()
	if err != nil {
		c.ui.Error(err.Error())
		return buildError
	}

	// Parse command flags
	flags := flag.NewFlagSet("build", flag.ContinueOnError)
	flags.StringVar(&main, "main", "main.go", "")
	flags.StringVar(&out, "out", "bin/"+repoName, "")
	flags.BoolVar(&all, "all", false, "")
	flags.Usage = func() { c.ui.Output(c.Help()) }
	if err := flags.Parse(args); err != nil {
		return buildFlagError
	}

	ctx, cancel := context.WithTimeout(context.Background(), buildTimeout)
	defer cancel()

	err = c.build(ctx, main, out)
	if err != nil {
		c.ui.Error(err.Error())
		return buildError
	}

	if all == true {
		err = c.buildAll(ctx, main, out)
		if err != nil {
			c.ui.Error(err.Error())
			return buildError
		}
	}

	return 0
}
