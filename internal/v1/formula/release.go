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

func (f *formula) ensure() (string, string, error) {
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

	f.Printf("⬇️  Pulling master branch ...")
	err = f.git.Pull()
	if err != nil {
		return "", "", err
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
	repo, branch, err := f.ensure()
	if err != nil {
		return err
	}

	current, next, err := f.versions(level)
	if err != nil {
		return err
	}

	f.Printf("➡️  Creating draft release %s ...", current.Version())
	release, err := f.github.CreateRelease(ctx, repo, service.ReleaseInput{
		Name:       current.Version(),
		TagName:    current.GitTag(),
		Target:     branch,
		Draft:      true,
		Prerelease: false,
	})

	if err != nil {
		return err
	}

	f.Printf("➡️  Creating/Updating change log ...")
	changelogText, err := f.changelog.Generate(ctx, current.GitTag())
	if err != nil {
		return err
	}

	f.Infof("▶️  Releasing current version %s ...", current.Version())

	commitMessage := fmt.Sprintf("Releasing %s", current.Version())
	err = f.git.Commit(commitMessage, f.spec.VersionFile, f.changelog.Filename())
	if err != nil {
		return err
	}

	err = f.git.Tag(current.GitTag())
	if err != nil {
		return err
	}

	// Building and uploading artifacts
	if f.spec.Release.Build {
		f.Printf("➡️  Building artifacts for release %s ...", release.Name)
		assets, err := f.CrossCompile(ctx)
		if err != nil {
			return err
		}

		for _, asset := range assets {
			f.Printf("⬆️  Uploading %s to release %s ...", asset, release.Name)
			err = f.github.UploadAssets(ctx, release, asset)
			if err != nil {
				return err
			}
			os.Remove(asset)
		}
	}

	f.Infof("▶️  Preparing next version %s ...", next.PreRelease())

	err = f.vmanager.Update(next.PreRelease())
	if err != nil {
		return err
	}

	commitMessage = fmt.Sprintf("Beginning %s [skip ci]", next.PreRelease())
	err = f.git.Commit(commitMessage, f.spec.VersionFile)
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

	f.Printf("⬆️  Pushing commits ...")
	err = f.git.Push(true)
	if err != nil {
		return err
	}

	f.Printf("⬆️  Publishing release %s ...", release.Name)
	release, err = f.github.EditRelease(ctx, repo, release.ID, service.ReleaseInput{
		Name:       current.Version(),
		TagName:    current.GitTag(),
		Target:     branch,
		Draft:      false,
		Prerelease: false,
		Body:       fmt.Sprintf("%s\n\n%s", comment, changelogText),
	})

	if err != nil {
		return err
	}

	return nil
}
