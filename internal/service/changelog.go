package service

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

const (
	changelogFilename = "CHANGELOG.md"
)

type (
	// Changelog is the interface for change log generation
	Changelog interface {
		Filename() string
		Generate(ctx context.Context, gitTag string) (string, error)
	}

	changelog struct {
		workDir     string
		githubToken string
	}
)

// NewChangelog creates a new instance of Changelog
func NewChangelog(workDir, githubToken string) Changelog {
	return &changelog{
		workDir:     workDir,
		githubToken: githubToken,
	}
}

func (c *changelog) Filename() string {
	return changelogFilename
}

func (c *changelog) Generate(ctx context.Context, gitTag string) (string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := exec.CommandContext(ctx,
		"github_changelog_generator",
		"--token", c.githubToken,
		"--no-filter-by-milestone",
		"--exclude-labels", "question,duplicate,invalid,wontfix",
		"--future-release", gitTag,
	)

	cmd.Dir = c.workDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%s: %s %s", err.Error(), stdout.String(), stderr.String())
	}

	file, err := os.Open(filepath.Join(c.workDir, changelogFilename))
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Regex for the start of current release
	startRE, err := regexp.Compile(fmt.Sprintf(`^## \[%s\]`, gitTag))
	if err != nil {
		return "", err
	}

	// Regex for the end of current release
	endRE, err := regexp.Compile(`^(##|\\\*)`)
	if err != nil {
		return "", err
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
		return "", err
	}

	changelog = strings.Trim(changelog, "\n")

	return changelog, nil
}
