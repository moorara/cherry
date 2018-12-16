package test

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/moorara/cherry/internal/exec"
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
		Cover(ctx context.Context) error
	}

	tester struct {
		workDir string
	}
)

// NewTester creates a new instance of test Tester
func NewTester(workDir string) Tester {
	return &tester{
		workDir: workDir,
	}
}

func (t *tester) getPackages(ctx context.Context) ([]string, error) {
	out, err := exec.Command(ctx, t.workDir, "go", "list", "./...")
	if err != nil {
		return nil, err
	}

	pkgs := strings.Split(out, "\n")

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

	// Run go test with cover mode
	_, err = exec.Command(ctx, t.workDir, "go", "test", "-covermode", atomicMode, "-coverprofile", tf.Name(), pkg)
	if err != nil {
		return err
	}

	// Get the coverage data
	out, err := exec.Command(ctx, t.workDir, "tail", "-n", "+2", tf.Name())
	if err != nil {
		return err
	}

	// Append the coverage data to the singleton coverage file
	if out != "" {
		_, err = cf.WriteString(out + "\n")
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *tester) Cover(ctx context.Context) error {
	reportpath := filepath.Join(t.workDir, reportPath)
	coverfile := filepath.Join(reportpath, coverFile)
	reportfile := filepath.Join(reportpath, reportFile)

	// Remove any prior coverage report
	err := os.RemoveAll(reportpath)
	if err != nil {
		return err
	}

	err = os.Mkdir(reportpath, os.ModePerm)
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
	_, err = exec.Command(ctx, t.workDir, "go", "tool", "cover", "-html", coverfile, "-o", reportfile)
	if err != nil {
		return err
	}

	return nil
}
