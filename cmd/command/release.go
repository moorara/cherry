package command

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"path/filepath"
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
		cli.Ui
		git.Git
		github.Github
		changelog.Changelog
		semver.Manager
		WorkDir      string
		VersionFile  string
		Repo, Branch string
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

		-patch:   create a patch version release                       (default: true)
		-minor:   create a minor version release                       (default: false)
		-major:   create a major version release                       (default: false)
		-comment: add a comment for the release
		-build:   build the artifacts and include them in the release  (default: false)
	
	Examples:

		cherry release
		cherry release -build
		cherry release -minor
		cherry release -minor -build
		cherry release -major
		cherry release -major -build
		cherry release -comment "release comment"
	`
)

// NewRelease create a new release command
func NewRelease(ui cli.Ui, workDir, githubToken string) (*Release, error) {
	if githubToken == "" {
		return nil, errors.New("github token is not set")
	}

	git := git.New(workDir)
	github := github.New(releaseTimeout, githubToken)
	changelog := changelog.New(workDir, githubToken)

	manager, err := semver.NewManager(filepath.Join(workDir, versionFile))
	if err != nil {
		return nil, err
	}

	owner, name, err := git.GetRepoName()
	if err != nil {
		return nil, err
	}
	repo := owner + "/" + name

	branch, err := git.GetBranchName()
	if err != nil {
		return nil, err
	}

	cmd := &Release{
		Ui:          ui,
		Git:         git,
		Github:      github,
		Changelog:   changelog,
		Manager:     manager,
		WorkDir:     workDir,
		VersionFile: versionFile,
		Repo:        repo,
		Branch:      branch,
	}

	return cmd, nil
}

func (c *Release) processVersions(rt releaseType) (semver.SemVer, semver.SemVer, error) {
	var empty, current, next semver.SemVer

	sv, err := c.Manager.Read()
	if err != nil {
		return empty, empty, err
	}

	switch rt {
	case patchRelease:
		current, next = sv.ReleasePatch()
	case minorRelease:
		current, next = sv.ReleaseMinor()
	case majorRelease:
		current, next = sv.ReleaseMajor()
	}

	err = c.Manager.Update(current.Version())
	if err != nil {
		return empty, empty, err
	}

	return current, next, nil
}

func (c *Release) release(ctx context.Context, rt releaseType, comment string, build bool) error {
	if c.Branch != "master" {
		return errors.New("release has to be run on master branch")
	}

	clean, err := c.Git.IsClean()
	if err != nil {
		return err
	}

	// This is to ensure that we do not commit any unwanted change while releasing
	if !clean {
		return errors.New("working directory is not clean and has uncommitted changes")
	}

	// Temporarily disable master branch protection
	c.Ui.Warn("🔓 Temporarily enabling push to master branch ...")
	err = c.Github.BranchProtectionForAdmin(ctx, c.Repo, c.Branch, false)
	if err != nil {
		return err
	}

	// Make sure we re-enable master branch protection
	defer func() {
		c.Ui.Warn("🔒 Re-disabling push to master branch ...")
		err = c.Github.BranchProtectionForAdmin(context.Background(), c.Repo, c.Branch, true)
		if err != nil {
			c.Ui.Error(err.Error())
		}
	}()

	// Release the current version and prepare the next version
	c.Ui.Info("🚀 Releasing current version ...")
	current, next, err := c.processVersions(rt)
	if err != nil {
		return err
	}

	// Create or update the change log
	changelogText, err := c.Changelog.Generate(ctx, current.GitTag())
	if err != nil {
		return err
	}

	commitMessage := fmt.Sprintf("Releasing %s", current.Version())
	err = c.Git.Commit(commitMessage, c.VersionFile, c.Changelog.Filename())
	if err != nil {
		return err
	}

	err = c.Git.Tag(current.GitTag())
	if err != nil {
		return err
	}

	err = c.Git.Push(true)
	if err != nil {
		return err
	}

	description := fmt.Sprintf("%s\n\n%s", comment, changelogText)
	release, err := c.Github.CreateRelease(ctx, c.Repo, c.Branch, current, description, false, false)
	if err != nil {
		return err
	}

	// Building and uploading artifacts
	if build {
		c.Ui.Info(fmt.Sprintf("🛠️ Building artifacts for release %s ...", release.Name))

		c.Ui.Info(fmt.Sprintf("📦 Uploading artifacts for release %s ...", release.Name))
	}

	// Prepare the version file for next version
	c.Ui.Info(fmt.Sprintf("✏️ Preparing next version %s ...", next.PreRelease()))
	err = c.Manager.Update(next.PreRelease())
	if err != nil {
		return err
	}

	commitMessage = fmt.Sprintf("Beginning %s", next.PreRelease())
	err = c.Git.Commit(commitMessage, c.VersionFile)
	if err != nil {
		return err
	}

	err = c.Git.Push(false)
	if err != nil {
		return err
	}

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
	var build bool

	// Parse command flags
	flags := flag.NewFlagSet("release", flag.ContinueOnError)
	flags.BoolVar(&patch, "patch", true, "")
	flags.BoolVar(&minor, "minor", false, "")
	flags.BoolVar(&major, "major", false, "")
	flags.StringVar(&comment, "comment", "", "")
	flags.BoolVar(&build, "build", false, "")
	flags.Usage = func() { c.Ui.Output(c.Help()) }
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

	err := c.release(ctx, rt, comment, build)
	if err != nil {
		c.Ui.Error(err.Error())
		return buildError
	}

	return 0
}
