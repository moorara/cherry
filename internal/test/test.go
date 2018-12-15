package test

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	atomicMode   = "atomic"
	atomicHeader = "mode: atomic\n"

	reportPath = "coverage"
	coverFile  = "cover.out"
	reportFile = "index.html"
)

type (
	// Tester is the interface for test runner
	Tester interface {
		Coverage(context.Context) error
	}

	tester struct {
		path string
	}
)

// NewTester creates a new instance of test Tester
func NewTester(path string) Tester {
	return &tester{
		path: path,
	}
}

func (t *tester) getPackages(ctx context.Context) ([]string, error) {
	cmd := exec.CommandContext(ctx, "go", "list", "./...")
	cmd.Dir = t.path
	out, err := cmd.Output()
	if err != nil {
		if e, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("%s. %s", e.Error(), string(e.Stderr))
		}
		return nil, err
	}

	pkgs := strings.Split(string(out), "\n")
	pkgs = pkgs[:len(pkgs)-1]

	return pkgs, nil
}

func (t *tester) testPackage(ctx context.Context, pkg, coverfile string) error {
	// Open the singleton coverage file
	cf, err := os.OpenFile(coverfile, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer cf.Close()

	// Create a temporary file for collecting cover output of each package
	tf, err := ioutil.TempFile("", "cover-*.out")
	if err != nil {
		return err
	}
	defer os.Remove(tf.Name())

	// Run go test in cover mode
	cmd := exec.CommandContext(ctx, "go", "test", "-covermode", atomicMode, "-coverprofile", tf.Name(), pkg)
	cmd.Dir = t.path
	_, err = cmd.Output()
	if err != nil {
		if e, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("%s. %s", e.Error(), string(e.Stderr))
		}
		return err
	}

	// Get the coverage data
	cmd = exec.CommandContext(ctx, "tail", "-n", "+2", tf.Name())
	cmd.Dir = t.path
	out, err := cmd.Output()
	if err != nil {
		if e, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("%s. %s", e.Error(), string(e.Stderr))
		}
		return err
	}

	// Append the coverage data to the singleton coverage file
	_, err = cf.WriteString(string(out))
	if err != nil {
		return err
	}

	return nil
}

func (t *tester) Coverage(ctx context.Context) error {
	coverpath := filepath.Join(t.path, reportPath)
	coverfile := filepath.Join(coverpath, coverFile)
	reportfile := filepath.Join(coverpath, reportFile)

	// Remove any prior coverage report
	err := os.RemoveAll(coverpath)
	if err != nil {
		return err
	}

	err = os.Mkdir(coverpath, os.ModePerm)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(coverfile, []byte(atomicHeader), 0644)
	if err != nil {
		return err
	}

	packages, err := t.getPackages(ctx)
	if err != nil {
		return err
	}

	for _, pkg := range packages {
		err = t.testPackage(ctx, pkg, coverfile)
		if err != nil {
			return err
		}
	}

	// Generate the singleton html coverage report for all packages
	cmd := exec.CommandContext(ctx, "go", "tool", "cover", "-html", coverfile, "-o", reportfile)
	cmd.Dir = t.path
	_, err = cmd.Output()
	if e, ok := err.(*exec.ExitError); ok {
		return fmt.Errorf("%s. %s", e.Error(), string(e.Stderr))
	}

	return nil
}
