package formula

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/mitchellh/cli"
	"github.com/moorara/cherry/internal/service/changelog"
	"github.com/moorara/cherry/internal/service/git"
	"github.com/moorara/cherry/internal/service/github"
	"github.com/moorara/cherry/internal/service/semver"
	"github.com/moorara/cherry/internal/v1/spec"
)

type (
	// Level specifies the level of a release (patch, minor, or major)
	ReleaseLevel int

	// Release is the interface for release formulas
	Release interface {
		Release(ctx context.Context, level ReleaseLevel, comment string) error
	}

	release struct {
		cli.Ui
		git.Git
		github.Github
		changelog.Changelog
		semver.Manager
		spec.Spec
		WorkDir string
	}
)

const (
	// PatchRelease releases a the patch component of a semantic version
	PatchRelease ReleaseLevel = iota
	// MinorRelease releases a the minor component of a semantic version
	MinorRelease
	// MajorRelease releases a the major component of a semantic version
	MajorRelease
)

const (
	githubTimeout = 30 * time.Second
)

// NewRelease creates a new instance of release formula
func NewRelease(ui cli.Ui, spec spec.Spec, workDir, githubToken string) (Release, error) {
	if githubToken == "" {
		return nil, errors.New("github token is not set")
	}

	git := git.New(workDir)
	github := github.New(githubTimeout, githubToken)
	changelog := changelog.New(workDir, githubToken)

	manager, err := semver.NewManager(filepath.Join(workDir, spec.VersionFile))
	if err != nil {
		return nil, err
	}

	return &release{
		Ui:        ui,
		Git:       git,
		Github:    github,
		Changelog: changelog,
		Manager:   manager,
		Spec:      spec,
		WorkDir:   workDir,
	}, nil
}

func (r *release) processVersions(level ReleaseLevel) (semver.SemVer, semver.SemVer, error) {
	var empty, current, next semver.SemVer

	sv, err := r.Manager.Read()
	if err != nil {
		return empty, empty, err
	}

	switch level {
	case PatchRelease:
		current, next = sv.ReleasePatch()
	case MinorRelease:
		current, next = sv.ReleaseMinor()
	case MajorRelease:
		current, next = sv.ReleaseMajor()
	}

	err = r.Manager.Update(current.Version())
	if err != nil {
		return empty, empty, err
	}

	return current, next, nil
}

func (r *release) Release(ctx context.Context, level ReleaseLevel, comment string) error {
	owner, name, err := r.Git.GetRepoName()
	if err != nil {
		return err
	}

	repo := owner + "/" + name

	branch, err := r.Git.GetBranchName()
	if err != nil {
		return err
	}

	if branch != "master" {
		return errors.New("release has to be run on master branch")
	}

	clean, err := r.Git.IsClean()
	if err != nil {
		return err
	}

	// This is to ensure that we do not commit any unwanted change while releasing
	if !clean {
		return errors.New("working directory is not clean and has uncommitted changes")
	}

	if r.Ui != nil {
		r.Ui.Warn("🔓 Temporarily enabling push to master branch ...")
	}

	err = r.Github.BranchProtectionForAdmin(ctx, repo, branch, false)
	if err != nil {
		return err
	}

	// Make sure we re-enable master branch protection
	defer func() {
		if r.Ui != nil {
			r.Ui.Warn("🔒 Re-disabling push to master branch ...")
		}

		err = r.Github.BranchProtectionForAdmin(context.Background(), repo, branch, true)
		if err != nil && r.Ui != nil {
			r.Ui.Error(err.Error())
		}
	}()

	if r.Ui != nil {
		r.Ui.Info("🚀 Releasing current version ...")
	}

	current, next, err := r.processVersions(level)
	if err != nil {
		return err
	}

	// Create or update the change log
	changelogText, err := r.Changelog.Generate(ctx, current.GitTag())
	if err != nil {
		return err
	}

	commitMessage := fmt.Sprintf("Releasing %s", current.Version())
	err = r.Git.Commit(commitMessage, r.Spec.VersionFile, r.Changelog.Filename())
	if err != nil {
		return err
	}

	err = r.Git.Tag(current.GitTag())
	if err != nil {
		return err
	}

	err = r.Git.Push(true)
	if err != nil {
		return err
	}

	description := fmt.Sprintf("%s\n\n%s", comment, changelogText)
	release, err := r.Github.CreateRelease(ctx, repo, branch, current, description, false, false)
	if err != nil {
		return err
	}

	// Building and uploading artifacts
	if r.Spec.Release.Build {
		if r.Ui != nil {
			r.Ui.Info(fmt.Sprintf("🛠️ Building artifacts for release %s ...", release.Name))
		}

		if r.Ui != nil {
			r.Ui.Info(fmt.Sprintf("📦 Uploading artifacts for release %s ...", release.Name))
		}
	}

	if r.Ui != nil {
		r.Ui.Info(fmt.Sprintf("✏️ Preparing next version %s ...", next.PreRelease()))
	}

	err = r.Manager.Update(next.PreRelease())
	if err != nil {
		return err
	}

	commitMessage = fmt.Sprintf("Beginning %s", next.PreRelease())
	err = r.Git.Commit(commitMessage, r.Spec.VersionFile)
	if err != nil {
		return err
	}

	err = r.Git.Push(false)
	if err != nil {
		return err
	}

	return nil
}
