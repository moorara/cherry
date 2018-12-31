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

	"github.com/moorara/cherry/internal/service/util"
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
)

func (f *formula) getBuildInfo(ctx context.Context) (*buildInfo, error) {
	info := new(buildInfo)

	data, err := ioutil.ReadFile(filepath.Join(f.WorkDir, f.Spec.VersionFile))
	if err != nil {
		return nil, err
	}

	info.Version = strings.Trim(string(data), "\n")

	info.Revision, err = f.Git.GetCommitSHA(true)
	if err != nil {
		return nil, err
	}

	info.Branch, err = f.Git.GetBranchName()
	if err != nil {
		return nil, err
	}

	info.GoVersion = runtime.Version()
	info.BuildTool = fmt.Sprintf("%s:%s", f.Spec.ToolName, f.Spec.ToolVersion)
	info.BuildTime = time.Now().UTC()

	return info, nil
}

func (f *formula) getLDFlags(ctx context.Context, info *buildInfo) (string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := exec.CommandContext(ctx, "go", "list", f.Spec.Build.VersionPackage)
	cmd.Dir = f.WorkDir
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

func (f *formula) Compile(ctx context.Context) error {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	info, err := f.getBuildInfo(ctx)
	if err != nil {
		return err
	}

	ldflags, err := f.getLDFlags(ctx, info)
	if err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, "go", "build", "-ldflags", ldflags, "-o", f.Spec.Build.BinaryFile, f.Spec.Build.MainFile)
	cmd.Dir = f.WorkDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %s", err.Error(), stderr.String())
	}

	if f.Ui != nil {
		f.Ui.Info(fmt.Sprintf("✅ %s", f.Spec.Build.BinaryFile))
	}

	return nil
}

func (f *formula) CrossCompile(ctx context.Context) error {
	info, err := f.getBuildInfo(ctx)
	if err != nil {
		return err
	}

	ldflags, err := f.getLDFlags(ctx, info)
	if err != nil {
		return err
	}

	for _, platform := range f.Spec.Build.Platforms {
		var stdout bytes.Buffer
		var stderr bytes.Buffer

		env := strings.Split(platform, "-")
		reset, err := util.SetEnvVars("GOOS", env[0], "GOARCH", env[1])
		if err != nil {
			return err
		}

		stdout.Reset()
		stderr.Reset()

		bin := fmt.Sprintf("%s-%s", f.Spec.Build.BinaryFile, platform)
		cmd := exec.CommandContext(ctx, "go", "build", "-ldflags", ldflags, "-o", bin, f.Spec.Build.MainFile)
		cmd.Dir = f.WorkDir
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("%s: %s", err.Error(), stderr.String())
		}

		// Restore environment variables
		reset()

		if f.Ui != nil {
			f.Ui.Info(fmt.Sprintf("✅ %s", bin))
		}
	}

	return nil
}
