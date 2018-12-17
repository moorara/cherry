package command

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/mitchellh/cli"
	"github.com/moorara/cherry/internal/exec"
	"github.com/moorara/cherry/internal/git"
)

const (
	buildTool      = "Cherry"
	versionFile    = "VERSION"
	versionPackage = "./cmd/version"

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

	vf, err := os.Open(filepath.Join(c.workDir, versionFile))
	if err != nil {
		return nil, err
	}
	defer vf.Close()

	sc := bufio.NewScanner(vf)
	for sc.Scan() {
		info.Version = sc.Text()
	}

	if err := sc.Err(); err != nil {
		return nil, err
	}

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
	vPkg, err := exec.Command(ctx, c.workDir, "go", "list", versionPackage)
	if err != nil {
		return "", err
	}

	versionFlag := fmt.Sprintf("-X %s.Version=%s", vPkg, info.Version)
	revisionFlag := fmt.Sprintf("-X %s.Revision=%s", vPkg, info.Revision)
	branchFlag := fmt.Sprintf("-X %s.Branch=%s", vPkg, info.Branch)
	goVersionFlag := fmt.Sprintf("-X %s.GoVersion=%s", vPkg, info.GoVersion)
	buildToolFlag := fmt.Sprintf("-X %s.BuildTool=%s", vPkg, info.BuildTool)
	buildTimeFlag := fmt.Sprintf("-X %s.BuildTime=%s", vPkg, info.BuildTime.Format(time.RFC3339Nano))

	ldflags := fmt.Sprintf("%s %s %s %s %s %s", versionFlag, revisionFlag, branchFlag, goVersionFlag, buildToolFlag, buildTimeFlag)

	return ldflags, err
}

func (c *Build) build(ctx context.Context, main, out string) error {
	info, err := c.getBuildInfo(ctx)
	if err != nil {
		return err
	}

	ldflags, err := c.getLDFlags(ctx, info)
	if err != nil {
		return err
	}

	_, err = exec.Command(ctx, c.workDir, "go", "build", "-ldflags", ldflags, "-o", out, main)
	if err != nil {
		return err
	}

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
		env := strings.Split(platform, "-")
		reset, err := exec.SetEnvVars("GOOS", env[0], "GOARCH", env[1])
		if err != nil {
			return err
		}

		out := fmt.Sprintf("%s-%s", outPrefix, platform)
		_, err = exec.Command(ctx, c.workDir, "go", "build", "-ldflags", ldflags, "-o", out, main)
		if err != nil {
			return err
		}

		// Restore environment variables
		reset()
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
