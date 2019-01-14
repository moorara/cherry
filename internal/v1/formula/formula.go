package formula

import (
	"context"
	"path/filepath"

	"github.com/mitchellh/cli"
	"github.com/moorara/cherry/internal/service"
	"github.com/moorara/cherry/internal/v1/spec"
)

type (
	// Formula is the interface for all available formulas
	Formula interface {
		Info(string)
		Warn(string)
		Error(string)

		Cover(context.Context) error
		Compile(context.Context) error
		CrossCompile(context.Context) ([]string, error)
		Release(ctx context.Context, level ReleaseLevel, comment string) error
	}

	formula struct {
		cli.Ui
		service.Git
		service.Github
		service.Changelog
		service.VersionManager
		*spec.Spec

		WorkDir     string
		GithubToken string
	}
)

// New creates a new instance of formula
func New(ui cli.Ui, spec *spec.Spec, workDir, githubToken string) (Formula, error) {
	git := service.NewGit(workDir)
	github := service.NewGithub(githubToken)
	changelog := service.NewChangelog(workDir, githubToken)
	vmanager := service.NewTextVersionManager(filepath.Join(workDir, spec.VersionFile))

	return &formula{
		Ui:             ui,
		Git:            git,
		Github:         github,
		Changelog:      changelog,
		VersionManager: vmanager,
		Spec:           spec,
		WorkDir:        workDir,
		GithubToken:    githubToken,
	}, nil
}

func (f *formula) Info(msg string) {
	if f.Ui != nil {
		f.Ui.Info(msg)
	}
}

func (f *formula) Warn(msg string) {
	if f.Ui != nil {
		f.Ui.Warn(msg)
	}
}

func (f *formula) Error(msg string) {
	if f.Ui != nil {
		f.Ui.Error(msg)
	}
}
