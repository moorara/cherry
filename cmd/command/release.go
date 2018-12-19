package command

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"github.com/mitchellh/cli"
	"github.com/moorara/cherry/internal/changelog"
	"github.com/moorara/cherry/internal/git"
	"github.com/moorara/cherry/internal/github"
	"github.com/moorara/cherry/internal/semver"
)

type (
	releaseType int

	// Release is the release CLI command
	Release struct {
		ui          cli.Ui
		workDir     string
		versionFile string
		git         git.Git
		github      github.Github
		githubToken string
		changelog   changelog.Changelog
	}
)

const (
	patchRelease releaseType = iota
	minorRelease
	majorRelease
)

const (
	releaseError     = 40
	releaseFlagError = 41
	releaseTimeout   = 1 * time.Minute
	releaseSynopsis  = `Create release`
	releaseHelp      = `
	Use this command for creating a new release.

	Flags:
		- major:   create a major version release  (default: false)
		- minor:   create a minor version release  (default: false)
		- patch:   create a patch version release  (default: true)
		- comment: add a comment for the release
	`
)

// NewRelease create a new release command
func NewRelease(ui cli.Ui, workDir, githubToken string) (*Release, error) {
	if githubToken == "" {
		return nil, errors.New("github token is not set")
	}

	git := git.New(workDir)
	github := github.New(releaseTimeout, githubToken)
	changelog := changelog.New(workDir)

	cmd := &Release{
		ui:          ui,
		workDir:     workDir,
		versionFile: versionFile,
		git:         git,
		github:      github,
		githubToken: githubToken,
		changelog:   changelog,
	}

	return cmd, nil
}

func (c *Release) getVersions(rt releaseType) (current semver.SemVer, next semver.SemVer, err error) {
	c.ui.Info("Releasing current version ...")

	var data []byte
	versionFilePath := filepath.Join(c.workDir, c.versionFile)

	data, err = ioutil.ReadFile(versionFilePath)
	if err != nil {
		return
	}

	version := strings.Trim(string(data), "\n")
	sv, err := semver.Parse(version)
	if err != nil {
		return
	}

	switch rt {
	case patchRelease:
		current, next = sv.ReleasePatch()
	case minorRelease:
		current, next = sv.ReleaseMinor()
	case majorRelease:
		current, next = sv.ReleaseMajor()
	}

	data = []byte(current.Version() + "\n")
	err = ioutil.WriteFile(versionFilePath, data, 0644)
	if err != nil {
		return
	}

	return
}

func (c *Release) release(ctx context.Context, rt releaseType, comment string) error {
	clean, err := c.git.IsClean()
	if err != nil {
		return err
	}

	// This is to ensure that we do not commit any unwanted change while releasing
	if !clean {
		return errors.New("working directory is not clean and has uncommitted changes")
	}

	owner, name, err := c.git.GetRepoName()
	if err != nil {
		return err
	}

	repo := owner + "/" + name
	branch := "master"

	// Temporarily disable master branch protection
	c.ui.Info("Temporarily enabling push to master branch ...")
	err = c.github.BranchProtectionForAdmin(ctx, repo, branch, true)
	if err != nil {
		return err
	}

	// Make sure we re-enable master branch protection
	defer func() {
		c.ui.Info("Re-disabling push to master branch ...")
		err = c.github.BranchProtectionForAdmin(ctx, repo, branch, false)
		if err != nil {
			c.ui.Error(err.Error())
		}
	}()

	// Release the current version and prepare the next version
	current, _, err := c.getVersions(rt)
	if err != nil {
		return err
	}

	currentVersion := "v" + current.Version()

	// Create or update the change log
	_, err = c.changelog.Generate(ctx, currentVersion)
	if err != nil {
		return err
	}

	commitMessage := fmt.Sprintf("Releasing %s", currentVersion)
	err = c.git.Commit(commitMessage, c.versionFile, c.changelog.Filename())
	if err != nil {
		return err
	}

	err = c.git.Tag(currentVersion)
	if err != nil {
		return err
	}

	/* err = c.git.Push(true)
	if err != nil {
		return err
	} */

	return nil
}

// Synopsis returns the short one-line synopsis of the command.
func (c *Release) Synopsis() string {
	return releaseSynopsis
}

// Help returns the long help text including usage, description, and list of flags for the command
func (c *Release) Help() string {
	return releaseHelp
}

// Run runs the actual command with the given CLI instance and command-line arguments
func (c *Release) Run(args []string) int {
	var patch, minor, major bool
	var comment string

	// Parse command flags
	flags := flag.NewFlagSet("release", flag.ContinueOnError)
	flags.BoolVar(&patch, "patch", true, "")
	flags.BoolVar(&minor, "minor", false, "")
	flags.BoolVar(&major, "major", false, "")
	flags.StringVar(&comment, "comment", "", "")
	flags.Usage = func() { c.ui.Output(c.Help()) }
	if err := flags.Parse(args); err != nil {
		return releaseFlagError
	}

	var rt releaseType
	if patch {
		rt = patchRelease
	} else if minor {
		rt = minorRelease
	} else if major {
		rt = majorRelease
	}

	ctx, cancel := context.WithTimeout(context.Background(), buildTimeout)
	defer cancel()

	err := c.release(ctx, rt, comment)
	if err != nil {
		c.ui.Error(err.Error())
		return buildError
	}

	return 0
}
