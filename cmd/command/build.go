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
	"text/template"
	"time"

	"github.com/mitchellh/cli"
	"github.com/moorara/cherry/cmd/version"
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
		cli.Ui
		git.Git
		WorkDir  string
		RepoName string
	}
)

const (
	buildTool = "cherry"

	buildError     = 20
	buildFlagError = 21
	buildTimeout   = 1 * time.Minute
	buildSynopsis  = `Build artifacts`
	buildHelp      = `
	Use this command for building artifacts.
	Currently, this command can only build Go applications.

	Flags:

		-all:  build the binary for all platforms  (default: false)
		-main: path to main.go file                (default: main.go)
		-out:  path for binary files               (default: bin/{{.RepoName}})

	Examples:

		cherry build
		cherry build -all
		cherry -main cmd/main.go -out build/app
	`
)

var (
	platforms = []string{"linux-386", "linux-amd64", "darwin-386", "darwin-amd64", "windows-386", "windows-amd64"}
)

// NewBuild create a new build command
func NewBuild(ui cli.Ui, workDir string) (*Build, error) {
	git := git.New(workDir)

	_, repoName, err := git.GetRepoName()
	if err != nil {
		ui.Error(err.Error())
		return nil, err
	}

	cmd := &Build{
		Ui:       ui,
		Git:      git,
		WorkDir:  workDir,
		RepoName: repoName,
	}

	return cmd, nil
}

func (c *Build) getBuildInfo(ctx context.Context) (*buildInfo, error) {
	info := new(buildInfo)

	data, err := ioutil.ReadFile(filepath.Join(c.WorkDir, versionFile))
	if err != nil {
		return nil, err
	}

	info.Version = strings.Trim(string(data), "\n")

	info.Revision, err = c.Git.GetCommitSHA(true)
	if err != nil {
		return nil, err
	}

	info.Branch, err = c.Git.GetBranchName()
	if err != nil {
		return nil, err
	}

	info.GoVersion = runtime.Version()
	info.BuildTool = fmt.Sprintf("%s-%s", buildTool, version.Version)
	info.BuildTime = time.Now().UTC()

	return info, nil
}

func (c *Build) getLDFlags(ctx context.Context, info *buildInfo) (string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := exec.CommandContext(ctx, "go", "list", versionPackage)
	cmd.Dir = c.WorkDir
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
	cmd.Dir = c.WorkDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %s", err.Error(), stderr.String())
	}

	c.Ui.Info(fmt.Sprintf("✅ %s", out))

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
		cmd.Dir = c.WorkDir
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("%s: %s", err.Error(), stderr.String())
		}

		// Restore environment variables
		reset()

		c.Ui.Info(fmt.Sprintf("✅ %s", out))
	}

	return nil
}

// Synopsis returns the short one-line synopsis of the command.
func (c *Build) Synopsis() string {
	return buildSynopsis
}

// Help returns the long help text including usage, description, and list of flags for the command
func (c *Build) Help() string {
	var buf bytes.Buffer
	t := template.Must(template.New("help").Parse(buildHelp))
	t.Execute(&buf, c)

	return buf.String()
}

// Run runs the actual command with the given CLI instance and command-line arguments
func (c *Build) Run(args []string) int {
	var all bool
	var main, out string
	defaultOut := "bin/" + c.RepoName

	// Parse command flags
	flags := flag.NewFlagSet("build", flag.ContinueOnError)
	flags.BoolVar(&all, "all", false, "")
	flags.StringVar(&main, "main", "main.go", "")
	flags.StringVar(&out, "out", defaultOut, "")
	flags.Usage = func() { c.Ui.Output(c.Help()) }
	if err := flags.Parse(args); err != nil {
		return buildFlagError
	}

	var err error

	ctx, cancel := context.WithTimeout(context.Background(), buildTimeout)
	defer cancel()

	if all {
		err = c.buildAll(ctx, main, out)
	} else {
		err = c.build(ctx, main, out)
	}

	if err != nil {
		c.Ui.Error(err.Error())
		return buildError
	}
	return 0
}
