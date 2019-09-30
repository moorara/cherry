package step

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GoVersion runs `go version` command.
type GoVersion struct {
	WorkDir string
	Result  struct {
		Version string
	}
}

// Dry is a dry run of the step.
func (s *GoVersion) Dry() error {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("go", "version")
	cmd.Dir = s.WorkDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %s", err.Error(), strings.Trim(stderr.String(), "\n"))
	}

	return nil
}

// Run executes the step.
func (s *GoVersion) Run() error {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("go", "version")
	cmd.Dir = s.WorkDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %s", err.Error(), strings.Trim(stderr.String(), "\n"))
	}

	s.Result.Version = stdout.String()

	return nil
}

// Revert reverts back an executed step.
func (s *GoVersion) Revert() error {
	return nil
}

// GoList runs `go list ...` command.
type GoList struct {
	WorkDir string
	Package string
	Result  struct {
		PackagePath string
	}
}

// Dry is a dry run of the step.
func (s *GoList) Dry() error {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("go", "list", s.Package)
	cmd.Dir = s.WorkDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %s", err.Error(), strings.Trim(stderr.String(), "\n"))
	}

	return nil
}

// Run executes the step.
func (s *GoList) Run() error {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("go", "list", s.Package)
	cmd.Dir = s.WorkDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %s", err.Error(), strings.Trim(stderr.String(), "\n"))
	}

	s.Result.PackagePath = strings.Trim(stdout.String(), "\n")

	return nil
}

// Revert reverts back an executed step.
func (s *GoList) Revert() error {
	return nil
}

// GoBuild runs `go build ...` command.
type GoBuild struct {
	WorkDir    string
	Ctx        context.Context
	LDFlags    string
	MainFile   string
	BinaryFile string
	Platforms  []string
	Result     struct {
		Binaries []string
	}
}

func (s *GoBuild) build(binaryFile string) error {
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
	cmd := exec.CommandContext(s.Ctx, "go", args...)
	cmd.Dir = s.WorkDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %s", err.Error(), strings.Trim(stderr.String(), "\n"))
	}

	s.Result.Binaries = append(s.Result.Binaries, binaryFile)

	return nil
}

// Dry is a dry run of the step.
func (s *GoBuild) Dry() error {
	s.Result.Binaries = []string{}

	dir, err := ioutil.TempDir("", "cherry-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	binaryFile := filepath.Join(dir, s.BinaryFile)
	err = s.build(binaryFile)
	if err != nil {
		return err
	}

	return nil
}

// Run executes the step.
func (s *GoBuild) Run() error {
	s.Result.Binaries = []string{}

	if len(s.Platforms) == 0 {
		return s.build(s.BinaryFile)
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
		err := s.build(binaryFile)
		if err != nil {
			return err
		}
	}

	return nil
}

// Revert reverts back an executed step.
func (s *GoBuild) Revert() error {
	for _, binary := range s.Result.Binaries {
		if err := os.Remove(binary); err != nil {
			return err
		}
	}

	return nil
}
