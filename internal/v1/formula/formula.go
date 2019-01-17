package formula

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/mitchellh/cli"
	"github.com/moorara/cherry/internal/service"
	"github.com/moorara/cherry/internal/v1/spec"
)

type (
	// Formula is the interface for all available formulas
	Formula interface {
		Printf(string, ...interface{})
		Infof(string, ...interface{})
		Warnf(string, ...interface{})
		Errorf(string, ...interface{})

		Cover(context.Context) error
		Compile(context.Context) error
		CrossCompile(context.Context) ([]string, error)
		Release(ctx context.Context, level ReleaseLevel, comment string) error
	}

	formula struct {
		workDir     string
		githubToken string
		spec        *spec.Spec
		ui          cli.Ui
		git         service.Git
		github      service.Github
		changelog   service.Changelog
		vmanager    service.VersionManager
	}
)

// New creates a new instance of formula
func New(workDir, githubToken string, spec *spec.Spec, ui cli.Ui) (Formula, error) {
	git := service.NewGit(workDir)
	github := service.NewGithub(githubToken)
	changelog := service.NewChangelog(workDir, githubToken)
	vmanager := service.NewTextVersionManager(filepath.Join(workDir, spec.VersionFile))

	return &formula{
		workDir:     workDir,
		githubToken: githubToken,
		spec:        spec,
		ui:          ui,
		git:         git,
		github:      github,
		changelog:   changelog,
		vmanager:    vmanager,
	}, nil
}

func (f *formula) Printf(msg string, args ...interface{}) {
	if f.ui != nil {
		f.ui.Output(fmt.Sprintf(msg, args...))
	}
}

func (f *formula) Infof(msg string, args ...interface{}) {
	if f.ui != nil {
		f.ui.Info(fmt.Sprintf(msg, args...))
	}
}

func (f *formula) Warnf(msg string, args ...interface{}) {
	if f.ui != nil {
		f.ui.Warn(fmt.Sprintf(msg, args...))
	}
}

func (f *formula) Errorf(msg string, args ...interface{}) {
	if f.ui != nil {
		f.ui.Error(fmt.Sprintf(msg, args...))
	}
}
