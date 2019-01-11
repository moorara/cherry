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
	data, err := ioutil.ReadFile(filepath.Join(f.WorkDir, f.Spec.VersionFile))
	if err != nil {
		return nil, err
	}
	version := strings.Trim(string(data), "\n")

	commit, err := f.Git.GetHEAD()
	if err != nil {
		return nil, err
	}

	branch, err := f.Git.GetBranch()
	if err != nil {
		return nil, err
	}

	info := &buildInfo{
		Version:   version,
		Revision:  commit.ShortSHA,
		Branch:    branch.Name,
		GoVersion: runtime.Version(),
		BuildTool: fmt.Sprintf("%s@%s", f.Spec.ToolName, f.Spec.ToolVersion),
		BuildTime: time.Now().UTC(),
	}

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

	f.Info(fmt.Sprintf("✅ %s", f.Spec.Build.BinaryFile))

	return nil
}

func (f *formula) CrossCompile(ctx context.Context) ([]string, error) {
	info, err := f.getBuildInfo(ctx)
	if err != nil {
		return nil, err
	}

	ldflags, err := f.getLDFlags(ctx, info)
	if err != nil {
		return nil, err
	}

	artifacts := make([]string, len(f.Spec.Build.Platforms))

	for i, platform := range f.Spec.Build.Platforms {
		var stdout bytes.Buffer
		var stderr bytes.Buffer

		env := strings.Split(platform, "-")
		reset, err := util.SetEnvVars("GOOS", env[0], "GOARCH", env[1])
		if err != nil {
			return nil, err
		}

		stdout.Reset()
		stderr.Reset()

		bin := fmt.Sprintf("%s-%s", f.Spec.Build.BinaryFile, platform)
		cmd := exec.CommandContext(ctx, "go", "build", "-ldflags", ldflags, "-o", bin, f.Spec.Build.MainFile)
		cmd.Dir = f.WorkDir
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("%s: %s", err.Error(), stderr.String())
		}

		artifacts[i] = bin
		f.Info(fmt.Sprintf("✅ %s", bin))

		// Restore environment variables
		reset()
	}

	return artifacts, nil
}
