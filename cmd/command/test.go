package command

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/mitchellh/cli"
)

type (
	// Test is the test CLI command
	Test struct {
		ui      cli.Ui
		workDir string
	}
)

const (
	atomicMode   = "atomic"
	atomicHeader = "mode: atomic\n"
	reportPath   = "coverage"
	coverFile    = "cover.out"
	reportFile   = "index.html"

	testError     = 40
	testFlagError = 41
	testTimeout   = 1 * time.Minute
	testSynopsis  = `Run tests`
	testHelp      = `
	Use this command for running unit tests and generating coverage report.
	`
)

// NewTest create a new test command
func NewTest(ui cli.Ui, workDir string) (*Test, error) {
	cmd := &Test{
		ui:      ui,
		workDir: workDir,
	}

	return cmd, nil
}

func (c *Test) getPackages(ctx context.Context) ([]string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := exec.CommandContext(ctx, "go", "list", "./...")
	cmd.Dir = c.workDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("%s: %s", err.Error(), stderr.String())
	}

	pkgs := strings.Split(stdout.String(), "\n")

	return pkgs, nil
}

func (c *Test) testPackage(ctx context.Context, pkg, coverfile string) error {
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
	cmd.Dir = c.workDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %s", err.Error(), stderr.String())
	}

	testOutput := strings.Trim(stdout.String(), "\n")
	c.ui.Output(testOutput)

	stdout.Reset()
	stderr.Reset()

	// Get the coverage data
	cmd = exec.CommandContext(ctx, "tail", "-n", "+2", tf.Name())
	cmd.Dir = c.workDir
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

func (c *Test) cover(ctx context.Context) error {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	reportpath := filepath.Join(c.workDir, reportPath)
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

	packages, err := c.getPackages(ctx)
	if err != nil {
		return err
	}

	for _, pkg := range packages {
		err = c.testPackage(ctx, pkg, coverfile)
		if err != nil {
			return err
		}
	}

	// Generate the singleton html coverage report for all packages
	cmd := exec.CommandContext(ctx, "go", "tool", "cover", "-html", coverfile, "-o", reportfile)
	cmd.Dir = c.workDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %s", err.Error(), stderr.String())
	}

	return nil
}

// Synopsis returns the short one-line synopsis of the command.
func (c *Test) Synopsis() string {
	return testSynopsis
}

// Help returns the long help text including usage, description, and list of flags for the command
func (c *Test) Help() string {
	return testHelp
}

// Run runs the actual command with the given CLI instance and command-line arguments
func (c *Test) Run(args []string) int {
	// Parse command flags
	flags := flag.NewFlagSet("test", flag.ContinueOnError)
	flags.Usage = func() { c.ui.Output(c.Help()) }
	if err := flags.Parse(args); err != nil {
		return testFlagError
	}

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	err := c.cover(ctx)
	if err != nil {
		c.ui.Error(err.Error())
		return testError
	}

	return 0
}
