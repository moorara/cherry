package build

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/moorara/cherry/internal/exec"
)

const (
	buildTool      = "Cherry"
	versionFile    = "VERSION"
	versionPackage = "./cmd/version"
)

var (
	platforms = []string{"linux-386", "linux-amd64", "darwin-386", "darwin-amd64", "windows-386", "windows-amd64"}
)

type (
	// Builder is the interface for building go applications
	Builder interface {
		Build(ctx context.Context, main, out string) error
		BuildAll(ctx context.Context, main, outPrefix string) error
	}

	builder struct {
		workDir string
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

// NewBuilder creates a new instance of Builder
func NewBuilder(workDir string) Builder {
	return &builder{
		workDir: workDir,
	}
}

func (b *builder) getBuildInfo(ctx context.Context) (*buildInfo, error) {
	info := new(buildInfo)

	vf, err := os.Open(filepath.Join(b.workDir, versionFile))
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

	info.Revision, err = exec.Command(ctx, b.workDir, "git", "rev-parse", "--short", "HEAD")
	if err != nil {
		return nil, err
	}

	info.Branch, err = exec.Command(ctx, b.workDir, "git", "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return nil, err
	}

	info.GoVersion = runtime.Version()
	info.BuildTool = buildTool
	info.BuildTime = time.Now().UTC()

	return info, nil
}

func (b *builder) getLDFlags(ctx context.Context, info *buildInfo) (string, error) {
	vPkg, err := exec.Command(ctx, b.workDir, "go", "list", versionPackage)
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

func (b *builder) Build(ctx context.Context, main, out string) error {
	info, err := b.getBuildInfo(ctx)
	if err != nil {
		return err
	}

	ldflags, err := b.getLDFlags(ctx, info)
	if err != nil {
		return err
	}

	_, err = exec.Command(ctx, b.workDir, "go", "build", "-ldflags", ldflags, "-o", out, main)
	if err != nil {
		return err
	}

	return nil
}

func (b *builder) BuildAll(ctx context.Context, main, outPrefix string) error {
	info, err := b.getBuildInfo(ctx)
	if err != nil {
		return err
	}

	ldflags, err := b.getLDFlags(ctx, info)
	if err != nil {
		return err
	}

	for _, platform := range platforms {
		env := strings.Split(platform, "-")
		os.Setenv("GOOS", env[0])
		os.Setenv("GOARCH", env[1])

		out := fmt.Sprintf("%s-%s", outPrefix, platform)

		_, err = exec.Command(ctx, b.workDir, "go", "build", "-ldflags", ldflags, "-o", out, main)
		if err != nil {
			return err
		}
	}

	os.Unsetenv("GOOS")
	os.Unsetenv("GOARCH")

	return nil
}
