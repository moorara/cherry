package step

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

const changelogFilename = "CHANGELOG.md"

// ChangelogGenerate runs `github_changelog_generator` Ruby gem!
type ChangelogGenerate struct {
	Mock        Step
	WorkDir     string
	GitHubToken string
	Repo        string
	Tag         string
	Result      struct {
		Filename  string
		Changelog string
	}
}

// Dry is a dry run of the step
func (s *ChangelogGenerate) Dry(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Dry(ctx)
	}

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "github_changelog_generator", "--version")
	cmd.Dir = s.WorkDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf(
			"ChangelogGenerate.Dry: %s %s %s",
			err.Error(),
			strings.Trim(stdout.String(), "\n"),
			strings.Trim(stderr.String(), "\n"),
		)
	}

	s.Result.Filename = changelogFilename

	return nil
}

// Run executes the step
func (s *ChangelogGenerate) Run(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Run(ctx)
	}

	var stdout, stderr bytes.Buffer

	cmd := exec.CommandContext(ctx,
		"github_changelog_generator",
		"--token", s.GitHubToken,
		"--no-filter-by-milestone",
		"--exclude-labels", "question,duplicate,invalid,wontfix",
		"--future-release", s.Tag,
		s.Repo,
	)

	cmd.Dir = s.WorkDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf(
			"ChangelogGenerate.Run: %s %s %s",
			err.Error(),
			strings.Trim(stdout.String(), "\n"),
			strings.Trim(stderr.String(), "\n"),
		)
	}

	file, err := os.Open(filepath.Join(s.WorkDir, changelogFilename))
	if err != nil {
		return err
	}
	defer file.Close()

	// Regex for the start of current release
	startRE, err := regexp.Compile(fmt.Sprintf(`^## \[%s\]`, s.Tag))
	if err != nil {
		return err
	}

	// Regex for the end of current release
	endRE, err := regexp.Compile(`^(##|\\\*)`)
	if err != nil {
		return err
	}

	var saveText bool
	var changelog string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if startRE.MatchString(line) {
			saveText = true
			continue
		} else if endRE.MatchString(line) {
			break
		}

		if saveText {
			changelog += line + "\n"
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	s.Result.Filename = changelogFilename
	s.Result.Changelog = strings.Trim(changelog, "\n")

	return nil
}

// Revert reverts back an executed step
func (s *ChangelogGenerate) Revert(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Revert(ctx)
	}

	// TODO: how to revert?
	return nil
}
