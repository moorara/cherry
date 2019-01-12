package formula

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/moorara/cherry/internal/service"
)

type (
	// ReleaseLevel specifies the level of a release (patch, minor, or major)
	ReleaseLevel int
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

func (f *formula) precheck() (string, string, error) {
	if f.GithubToken == "" {
		return "", "", errors.New("github token is not set")
	}

	repo, err := f.Git.GetRepo()
	if err != nil {
		return "", "", err
	}

	branch, err := f.Git.GetBranch()
	if err != nil {
		return "", "", err
	}

	if branch.Name != "master" {
		return "", "", errors.New("release has to be run on master branch")
	}

	clean, err := f.Git.IsClean()
	if err != nil {
		return "", "", err
	}

	// This is to ensure that we do not commit any unwanted change while releasing
	if !clean {
		return "", "", errors.New("working directory is not clean and has uncommitted changes")
	}

	return repo.Path(), branch.Name, nil
}

func (f *formula) versions(level ReleaseLevel) (service.SemVer, service.SemVer, error) {
	var empty, current, next service.SemVer

	sv, err := f.VersionManager.Read()
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

	err = f.VersionManager.Update(current.Version())
	if err != nil {
		return empty, empty, err
	}

	return current, next, nil
}

func (f *formula) Release(ctx context.Context, level ReleaseLevel, comment string) error {
	repo, branch, err := f.precheck()
	if err != nil {
		return err
	}

	f.Warn("🔓 Temporarily enabling push to master branch ...")

	err = f.Github.BranchProtectionForAdmin(ctx, repo, branch, false)
	if err != nil {
		return err
	}

	// Make sure we re-enable master branch protection
	defer func() {
		f.Warn("🔒 Re-disabling push to master branch ...")
		err = f.Github.BranchProtectionForAdmin(context.Background(), repo, branch, true)
		if err != nil {
			f.Error(err.Error())
		}
	}()

	current, next, err := f.versions(level)
	if err != nil {
		return err
	}

	f.Info(fmt.Sprintf("🚀 Releasing current version %s ...", current.Version()))

	// Create or update the change log
	changelogText, err := f.Changelog.Generate(ctx, current.GitTag())
	if err != nil {
		return err
	}

	commitMessage := fmt.Sprintf("Releasing %s", current.Version())
	err = f.Git.Commit(commitMessage, f.Spec.VersionFile, f.Changelog.Filename())
	if err != nil {
		return err
	}

	err = f.Git.Tag(current.GitTag())
	if err != nil {
		return err
	}

	err = f.Git.Push(true)
	if err != nil {
		return err
	}

	description := fmt.Sprintf("%s\n\n%s", comment, changelogText)
	release, err := f.Github.CreateRelease(ctx, repo, branch, current, description, false, false)
	if err != nil {
		return err
	}

	// Building and uploading artifacts
	if f.Spec.Release.Build {
		f.Info(fmt.Sprintf("🛠️  Building artifacts for release %s ...", release.Name))
		assets, err := f.CrossCompile(ctx)
		if err != nil {
			// We don't break the release process if we cannot build artifacts
			f.Error(fmt.Sprintf("🔴 Error on building artifacts: %s", err))
		} else {
			f.Info(fmt.Sprintf("📦 Uploading artifacts for release %s ...", release.Name))
			err = f.Github.UploadAssets(ctx, repo, current, assets)
			if err != nil {
				// We don't break the release process if we cannot upload artifacts
				f.Error(fmt.Sprintf("🔴 Error on uploading artifacts: %s", err))
			}
		}

		f.Info("🧹 Cleaning up artifacts ...")
		for _, asset := range assets {
			err := os.Remove(asset)
			if err != nil {
				f.Warn(fmt.Sprintf("🔴 Error on removing %s: %s", asset, err))
			}
		}
	}

	f.Info(fmt.Sprintf("✏️  Preparing next version %s ...", next.PreRelease()))

	err = f.VersionManager.Update(next.PreRelease())
	if err != nil {
		return err
	}

	commitMessage = fmt.Sprintf("Beginning %s", next.PreRelease())
	err = f.Git.Commit(commitMessage, f.Spec.VersionFile)
	if err != nil {
		return err
	}

	err = f.Git.Push(false)
	if err != nil {
		return err
	}

	return nil
}
