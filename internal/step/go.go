package step

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// GoVersion runs `go version` command.
type GoVersion struct {
	Mock    Step
	WorkDir string
	Result  struct {
		Version string
	}
}

// Dry is a dry run of the step.
func (s *GoVersion) Dry(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Dry(ctx)
	}

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "go", "version")
	cmd.Dir = s.WorkDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("GoVersion.Dry: %s %s", err.Error(), strings.Trim(stderr.String(), "\n"))
	}

	return nil
}

// Run executes the step.
func (s *GoVersion) Run(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Run(ctx)
	}

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "go", "version")
	cmd.Dir = s.WorkDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("GoVersion.Run: %s %s", err.Error(), strings.Trim(stderr.String(), "\n"))
	}

	// Get the version of Go compiler
	re := regexp.MustCompile(`go\d+\.\d+(\.\d+)?`)
	goVersion := re.FindString(stdout.String())

	s.Result.Version = goVersion

	return nil
}

// Revert reverts back an executed step.
func (s *GoVersion) Revert(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Revert(ctx)
	}

	return nil
}

// GoList runs `go list ...` command.
type GoList struct {
	Mock    Step
	WorkDir string
	Package string
	Result  struct {
		PackagePath string
	}
}

// Dry is a dry run of the step.
func (s *GoList) Dry(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Dry(ctx)
	}

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "go", "list", s.Package)
	cmd.Dir = s.WorkDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("GoList.Dry: %s %s", err.Error(), strings.Trim(stderr.String(), "\n"))
	}

	return nil
}

// Run executes the step.
func (s *GoList) Run(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Run(ctx)
	}

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "go", "list", s.Package)
	cmd.Dir = s.WorkDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("GoList.Run: %s %s", err.Error(), strings.Trim(stderr.String(), "\n"))
	}

	s.Result.PackagePath = strings.Trim(stdout.String(), "\n")

	return nil
}

// Revert reverts back an executed step.
func (s *GoList) Revert(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Revert(ctx)
	}

	return nil
}

// GoBuild runs `go build ...` command.
type GoBuild struct {
	Mock       Step
	WorkDir    string
	LDFlags    string
	MainFile   string
	BinaryFile string
	Platforms  []string
	Result     struct {
		Binaries []string
	}
}

func (s *GoBuild) build(ctx context.Context, binaryFile string) error {
	if s.MainFile == "" {
		s.MainFile = "main.go"
	}

	args := []string{"build"}
	if s.LDFlags != "" {
		args = append(args, "-ldflags", s.LDFlags)
	}
	if binaryFile != "" {
		args = append(args, "-o", binaryFile)
	}
	args = append(args, s.MainFile)

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = s.WorkDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s %s", err.Error(), strings.Trim(stderr.String(), "\n"))
	}

	s.Result.Binaries = append(s.Result.Binaries, binaryFile)

	return nil
}

// Dry is a dry run of the step.
func (s *GoBuild) Dry(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Dry(ctx)
	}

	dir, err := ioutil.TempDir("", "cherry-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	s.Result.Binaries = []string{}
	binaryFile := filepath.Join(dir, s.BinaryFile)
	err = s.build(ctx, binaryFile)
	if err != nil {
		return fmt.Errorf("GoBuild.Dry: %s", err)
	}

	return nil
}

// Run executes the step.
func (s *GoBuild) Run(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Run(ctx)
	}

	s.Result.Binaries = []string{}

	if len(s.Platforms) == 0 {
		return s.build(ctx, s.BinaryFile)
	}

	// Cross-Compile

	defer os.Unsetenv("GOOS")
	defer os.Unsetenv("GOARCH")

	for _, platform := range s.Platforms {
		env := strings.Split(platform, "-")

		if err := os.Setenv("GOOS", env[0]); err != nil {
			return err
		}

		if err := os.Setenv("GOARCH", env[1]); err != nil {
			return err
		}

		binaryFile := fmt.Sprintf("%s-%s", s.BinaryFile, platform)
		err := s.build(ctx, binaryFile)
		if err != nil {
			return fmt.Errorf("GoBuild.Run: %s", err)
		}
	}

	return nil
}

// Revert reverts back an executed step.
func (s *GoBuild) Revert(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Revert(ctx)
	}

	for _, binary := range s.Result.Binaries {
		if err := os.Remove(binary); err != nil {
			return fmt.Errorf("GoBuild.Revert: %s", err)
		}
	}

	return nil
}
