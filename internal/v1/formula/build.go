package formula

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/mitchellh/cli"
	"github.com/moorara/cherry/internal/service/git"
	"github.com/moorara/cherry/internal/service/util"
	"github.com/moorara/cherry/internal/v1/spec"
)

type (
	// Build is the interface for build formulas
	Build interface {
		Compile(ctx context.Context) error
		CrossCompile(ctx context.Context) error
	}

	build struct {
		cli.Ui
		git.Git
		spec.Spec
		WorkDir string
	}

	buildInfo struct {
		Version   string
		Revision  string
		Branch    string
		GoVersion string
		BuildTool string
		BuildTime time.Time
	}
)

// NewBuild creates a new instance of build formula
func NewBuild(ui cli.Ui, spec spec.Spec, workDir string) Build {
	git := git.New(workDir)

	return &build{
		Ui:      ui,
		Git:     git,
		Spec:    spec,
		WorkDir: workDir,
	}
}

func (b *build) getBuildInfo(ctx context.Context) (*buildInfo, error) {
	info := new(buildInfo)

	data, err := ioutil.ReadFile(filepath.Join(b.WorkDir, b.Spec.VersionFile))
	if err != nil {
		return nil, err
	}

	info.Version = strings.Trim(string(data), "\n")

	info.Revision, err = b.Git.GetCommitSHA(true)
	if err != nil {
		return nil, err
	}

	info.Branch, err = b.Git.GetBranchName()
	if err != nil {
		return nil, err
	}

	info.GoVersion = runtime.Version()
	info.BuildTool = fmt.Sprintf("%s:%s", b.Spec.ToolName, b.Spec.ToolVersion)
	info.BuildTime = time.Now().UTC()

	return info, nil
}

func (b *build) getLDFlags(ctx context.Context, info *buildInfo) (string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := exec.CommandContext(ctx, "go", "list", b.Spec.Build.VersionPackage)
	cmd.Dir = b.WorkDir
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

func (b *build) Compile(ctx context.Context) error {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	info, err := b.getBuildInfo(ctx)
	if err != nil {
		return err
	}

	ldflags, err := b.getLDFlags(ctx, info)
	if err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, "go", "build", "-ldflags", ldflags, "-o", b.Spec.Build.BinaryFile, b.Spec.Build.MainFile)
	cmd.Dir = b.WorkDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %s", err.Error(), stderr.String())
	}

	if b.Ui != nil {
		b.Ui.Info(fmt.Sprintf("✅ %s", b.Spec.Build.BinaryFile))
	}

	return nil
}

func (b *build) CrossCompile(ctx context.Context) error {
	info, err := b.getBuildInfo(ctx)
	if err != nil {
		return err
	}

	ldflags, err := b.getLDFlags(ctx, info)
	if err != nil {
		return err
	}

	for _, platform := range b.Spec.Build.Platforms {
		var stdout bytes.Buffer
		var stderr bytes.Buffer

		env := strings.Split(platform, "-")
		reset, err := util.SetEnvVars("GOOS", env[0], "GOARCH", env[1])
		if err != nil {
			return err
		}

		stdout.Reset()
		stderr.Reset()

		bin := fmt.Sprintf("%s-%s", b.Spec.Build.BinaryFile, platform)
		cmd := exec.CommandContext(ctx, "go", "build", "-ldflags", ldflags, "-o", bin, b.Spec.Build.MainFile)
		cmd.Dir = b.WorkDir
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("%s: %s", err.Error(), stderr.String())
		}

		// Restore environment variables
		reset()

		if b.Ui != nil {
			b.Ui.Info(fmt.Sprintf("✅ %s", bin))
		}
	}

	return nil
}
