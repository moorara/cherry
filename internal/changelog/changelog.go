package changelog

import (
	"context"

	"github.com/moorara/cherry/internal/exec"
)

const (
	changelogFilename = "CHANGELOG.md"
)

type (
	// Changelog is the interface for change log generation
	Changelog interface {
		Filename() string
		Generate(ctx context.Context, version string) (string, error)
	}

	changelog struct {
		workDir string
	}
)

// New creates a new instance of Changelog
func New(workDir string) Changelog {
	return &changelog{
		workDir: workDir,
	}
}

func (c *changelog) Filename() string {
	return changelogFilename
}

func (c *changelog) Generate(ctx context.Context, version string) (string, error) {
	_, err := exec.Command(ctx, c.workDir,
		"github_changelog_generator",
		"--no-filter-by-milestone",
		"--exclude-labels", "question,duplicate,invalid,wontfix",
		"--future-release", version,
	)

	// TODO: get difference!

	return "", err
}
