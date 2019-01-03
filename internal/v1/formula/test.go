package formula

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

const (
	atomicMode   = "atomic"
	atomicHeader = "mode: atomic\n"
	coverFile    = "cover.out"
	reportFile   = "index.html"
)

func (f *formula) getPackages(ctx context.Context) ([]string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := exec.CommandContext(ctx, "go", "list", "./...")
	cmd.Dir = f.WorkDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("%s: %s", err.Error(), stderr.String())
	}

	pkgs := strings.Split(stdout.String(), "\n")

	return pkgs, nil
}

func (f *formula) testPackage(ctx context.Context, pkg, coverfile string) error {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

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
	cmd := exec.CommandContext(ctx, "go", "test", "-covermode", atomicMode, "-coverprofile", tf.Name(), pkg)
	cmd.Dir = f.WorkDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	testOutput := strings.Trim(stdout.String(), "\n")

	if err != nil {
		f.Error(fmt.Sprintf("\n%s\n", testOutput))
		return fmt.Errorf("%s: %s", err.Error(), stderr.String())
	}

	f.Info(fmt.Sprintf("✅ %s", testOutput))

	stdout.Reset()
	stderr.Reset()

	// Get the coverage data
	cmd = exec.CommandContext(ctx, "tail", "-n", "+2", tf.Name())
	cmd.Dir = f.WorkDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %s", err.Error(), stderr.String())
	}

	// Append the coverage data to the singleton coverage file
	if out := stdout.String(); out != "" {
		_, err = cf.WriteString(out)
		if err != nil {
			return err
		}
	}

	return nil
}

func (f *formula) Cover(ctx context.Context) error {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	reportpath := filepath.Join(f.WorkDir, f.Spec.Test.ReportPath)
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

	packages, err := f.getPackages(ctx)
	if err != nil {
		return err
	}

	for _, pkg := range packages {
		err = f.testPackage(ctx, pkg, coverfile)
		if err != nil {
			return err
		}
	}

	// Generate the singleton html coverage report for all packages
	cmd := exec.CommandContext(ctx, "go", "tool", "cover", "-html", coverfile, "-o", reportfile)
	cmd.Dir = f.WorkDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %s", err.Error(), stderr.String())
	}

	return nil
}
