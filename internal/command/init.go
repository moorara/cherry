package command

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/mitchellh/cli"
)

const (
	initFlagErr        = 101
	initSpecFileErr    = 102
	initVersionFileErr = 103
	// initTimeout     = 10 * time.Second

	initSynopsis = `add cherry files`
	initHelp     = `
	Use this command for adding the files used by cherry.

	Examples:

		cherry init
	`

	initSpecFileContent = `version: "1.0"

language: go

build:
  cross_compile: false
  main_file: main.go
  version_package: ./version

release:
  build: false
`

	initVersionFileContent = `package version

var (
	// Version is the semantic version
	Version string

	// Commit is the SHA-1 of the git commit
	Commit string

	// Branch is the name of the git branch
	Branch string

	// GoVersion is the go compiler version
	GoVersion string

	// BuildTool contains the name and version of build tool
	BuildTool string

	// BuildTime is the time binary built
	BuildTime string
)
`
)

// initCommand implements cli.Command interface.
type initCommand struct {
	ui cli.Ui
}

// NewInitCommand creates an init command.
func NewInitCommand(ui cli.Ui) (cli.Command, error) {
	return &initCommand{
		ui: ui,
	}, nil
}

// Synopsis returns a short one-line synopsis of the command.
func (c *initCommand) Synopsis() string {
	return initSynopsis
}

// Help returns a long help text including usage, description, and list of flags for the command.
func (c *initCommand) Help() string {
	return initHelp
}

// Run runs the actual command with the given command-line arguments.
func (c *initCommand) Run(args []string) int {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	fs.Usage = func() {
		c.ui.Output(c.Help())
	}

	if err := fs.Parse(args); err != nil {
		return initFlagErr
	}

	// Adding Cherry spec file
	{
		specFilePath := filepath.Join(".", "cherry.yaml")

		specFileExist := false
		specFiles := []string{"cherry.yml", "cherry.yaml", "cherry.json"}
		for _, specFile := range specFiles {
			if _, err := os.Stat(filepath.Join(".", specFile)); err == nil {
				specFileExist = true
				break
			}
		}

		if !specFileExist {
			if err := ioutil.WriteFile(specFilePath, []byte(initSpecFileContent), 0644); err != nil {
				c.ui.Error(fmt.Sprintf("Error on writing spec file: %s", err))
				return initSpecFileErr
			}

			c.ui.Info(fmt.Sprintf("🍒 Spec file written to %s", specFilePath))
		}
	}

	// Adding Go version file
	{
		versionDirPath := filepath.Join(".", "version")
		versionFilePath := filepath.Join(versionDirPath, "version.go")

		if _, err := os.Stat(versionDirPath); os.IsNotExist(err) {
			if err := os.MkdirAll(versionDirPath, os.ModePerm); err != nil {
				c.ui.Error(fmt.Sprintf("Error on creating version package: %s", err))
				return initVersionFileErr
			}
		}

		if _, err := os.Stat(versionFilePath); os.IsNotExist(err) {
			if err := ioutil.WriteFile(versionFilePath, []byte(initVersionFileContent), 0644); err != nil {
				c.ui.Error(fmt.Sprintf("Error on writing version file: %s", err))
				return initVersionFileErr
			}

			c.ui.Info(fmt.Sprintf("🍒 Version file written to %s", versionFilePath))
		}
	}

	return 0
}
