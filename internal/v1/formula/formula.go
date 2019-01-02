package formula

import (
	"context"
	"path/filepath"

	"github.com/mitchellh/cli"
	"github.com/moorara/cherry/internal/service/changelog"
	"github.com/moorara/cherry/internal/service/git"
	"github.com/moorara/cherry/internal/service/github"
	"github.com/moorara/cherry/internal/service/semver"
	"github.com/moorara/cherry/internal/v1/spec"
)

type (
	// Formula is the interface for all available formulas
	Formula interface {
		Cover(ctx context.Context) error
		Compile(ctx context.Context) error
		CrossCompile(ctx context.Context) error
		Release(ctx context.Context, level ReleaseLevel, comment string) error
	}

	formula struct {
		cli.Ui
		git.Git
		github.Github
		changelog.Changelog
		semver.Manager
		*spec.Spec

		WorkDir     string
		GithubToken string
	}
)

// New creates a new instance of formula
func New(ui cli.Ui, spec *spec.Spec, workDir, githubToken string) (Formula, error) {
	git := git.New(workDir)
	github := github.New(githubTimeout, githubToken)
	changelog := changelog.New(workDir, githubToken)

	manager, err := semver.NewManager(filepath.Join(workDir, spec.VersionFile))
	if err != nil {
		return nil, err
	}

	return &formula{
		Ui:          ui,
		Git:         git,
		Github:      github,
		Changelog:   changelog,
		Manager:     manager,
		Spec:        spec,
		WorkDir:     workDir,
		GithubToken: githubToken,
	}, nil
}
