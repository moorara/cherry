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
)

func (f *formula) getBuildInfo(ctx context.Context) (*buildInfo, error) {
	data, err := ioutil.ReadFile(filepath.Join(f.workDir, f.spec.VersionFile))
	if err != nil {
		return nil, err
	}
	version := strings.Trim(string(data), "\n")

	commit, err := f.git.GetHEAD()
	if err != nil {
		return nil, err
	}

	branch, err := f.git.GetBranch()
	if err != nil {
		return nil, err
	}

	buildTool := f.spec.ToolName
	if f.spec.ToolVersion != "" {
		buildTool += "@" + f.spec.ToolVersion
	}

	info := &buildInfo{
		Version:   version,
		Revision:  commit.ShortSHA,
		Branch:    branch.Name,
		GoVersion: runtime.Version(),
		BuildTool: buildTool,
		BuildTime: time.Now().UTC(),
	}

	return info, nil
}

func (f *formula) getLDFlags(ctx context.Context, info *buildInfo) (string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := exec.CommandContext(ctx, "go", "list", f.spec.Build.VersionPackage)
	cmd.Dir = f.workDir
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

	cmd := exec.CommandContext(ctx, "go", "build", "-ldflags", ldflags, "-o", f.spec.Build.BinaryFile, f.spec.Build.MainFile)
	cmd.Dir = f.workDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %s", err.Error(), stderr.String())
	}

	f.Infof("🍒 %s", f.spec.Build.BinaryFile)

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

	artifacts := make([]string, len(f.spec.Build.Platforms))

	for i, platform := range f.spec.Build.Platforms {
		var stdout bytes.Buffer
		var stderr bytes.Buffer

		env := strings.Split(platform, "-")
		reset, err := util.SetEnvVars("GOOS", env[0], "GOARCH", env[1])
		if err != nil {
			return nil, err
		}

		stdout.Reset()
		stderr.Reset()

		bin := fmt.Sprintf("%s-%s", f.spec.Build.BinaryFile, platform)
		cmd := exec.CommandContext(ctx, "go", "build", "-ldflags", ldflags, "-o", bin, f.spec.Build.MainFile)
		cmd.Dir = f.workDir
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("%s: %s", err.Error(), stderr.String())
		}

		artifacts[i] = bin
		f.Infof("🍒 %s", bin)

		// Restore environment variables
		reset()
	}

	return artifacts, nil
}
