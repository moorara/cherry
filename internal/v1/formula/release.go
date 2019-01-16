package formula

import (
	"context"
	"errors"
	"fmt"
	"os"

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

func (f *formula) precheck() (string, string, error) {
	if f.githubToken == "" {
		return "", "", errors.New("github token is not set")
	}

	repo, err := f.git.GetRepo()
	if err != nil {
		return "", "", err
	}

	branch, err := f.git.GetBranch()
	if err != nil {
		return "", "", err
	}

	if branch.Name != "master" {
		return "", "", errors.New("release has to be run on master branch")
	}

	clean, err := f.git.IsClean()
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

	sv, err := f.vmanager.Read()
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

	err = f.vmanager.Update(current.Version())
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

	f.Printf("⬇️  Pulling master branch ...")
	err = f.git.Pull()
	if err != nil {
		return err
	}

	f.Warnf("🔓 Temporarily enabling push to master branch ...")
	err = f.github.BranchProtectionForAdmin(ctx, repo, branch, false)
	if err != nil {
		return err
	}

	// Make sure we re-enable master branch protection
	defer func() {
		f.Warnf("🔒 Re-disabling push to master branch ...")
		err = f.github.BranchProtectionForAdmin(context.Background(), repo, branch, true)
		if err != nil {
			f.Errorf(err.Error())
		}
	}()

	current, next, err := f.versions(level)
	if err != nil {
		return err
	}

	f.Infof("🚀 Releasing current version %s ...", current.Version())

	// Create or update the change log
	changelogText, err := f.changelog.Generate(ctx, current.GitTag())
	if err != nil {
		return err
	}

	commitMessage := fmt.Sprintf("Releasing %s", current.Version())
	err = f.git.Commit(commitMessage, f.spec.VersionFile, f.changelog.Filename())
	if err != nil {
		return err
	}

	err = f.git.Tag(current.GitTag())
	if err != nil {
		return err
	}

	err = f.git.Push(true)
	if err != nil {
		return err
	}

	description := fmt.Sprintf("%s\n\n%s", comment, changelogText)
	release, err := f.github.CreateRelease(ctx, repo, branch, current, description, false, false)
	if err != nil {
		return err
	}

	// Building and uploading artifacts
	if f.spec.Release.Build {
		f.Printf("🏗️ Building artifacts for release %s ...", release.Name)
		assets, err := f.CrossCompile(ctx)
		if err != nil {
			// We don't break the release process if we cannot build artifacts
			f.Errorf("🔴 Error on building artifacts: %s", err)
		} else {
			f.Printf("⬆️ Uploading artifacts for release %s ...", release.Name)
			err = f.github.UploadAssets(ctx, repo, current, assets)
			if err != nil {
				// We don't break the release process if we cannot upload artifacts
				f.Errorf("🔴 Error on uploading artifacts: %s", err)
			}
		}

		for _, asset := range assets {
			os.Remove(asset)
		}
	}

	f.Infof("✏️  Preparing next version %s ...", next.PreRelease())

	err = f.vmanager.Update(next.PreRelease())
	if err != nil {
		return err
	}

	commitMessage = fmt.Sprintf("Beginning %s [skip ci]", next.PreRelease())
	err = f.git.Commit(commitMessage, f.spec.VersionFile)
	if err != nil {
		return err
	}

	err = f.git.Push(false)
	if err != nil {
		return err
	}

	return nil
}
