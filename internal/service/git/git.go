package git

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

type (
	// Git is the interface for a git client
	Git interface {
		IsClean() (bool, error)
		GetRepoName() (string, string, error)
		GetBranchName() (string, error)
		GetCommitSHA(short bool) (string, error)
		Commit(message string, files ...string) error
		Tag(tag string) error
		Push(includeTags bool) error
	}

	git struct {
		workDir string
	}
)

// New creates a new git client
func New(workDir string) Git {
	return &git{
		workDir: workDir,
	}
}

// IsClean determines if the Git repo has any uncommitted changes
func (g *git) IsClean() (bool, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = g.workDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return false, fmt.Errorf("%s: %s", err.Error(), stderr.String())
	}

	return len(stdout.String()) == 0, nil
}

// GetRepoName returns the owner and name of Git repo
func (g *git) GetRepoName() (string, string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := exec.Command("git", "remote", "-v")
	cmd.Dir = g.workDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", "", fmt.Errorf("%s: %s", err.Error(), stderr.String())
	}

	// origin  git@github.com:USERNAME/REPOSITORY.git (push)     --> git@github.com:USERNAME/REPOSITORY.git
	// origin  https://github.com/USERNAME/REPOSITORY.git (push) --> https://github.com/USERNAME/REPOSITORY.git
	re := regexp.MustCompile(`origin[[:blank:]]+(.*)[[:blank:]]\(push\)`)
	subs := re.FindStringSubmatch(string(stdout.String()))
	if len(subs) != 2 {
		return "", "", errors.New("failed to get git repository url")
	}

	gitURL := subs[1]

	// git@github.com:USERNAME/REPOSITORY.git     --> USERNAME/REPOSITORY.git
	// https://github.com/USERNAME/REPOSITORY.git --> USERNAME/REPOSITORY.git
	re = regexp.MustCompile(`(git@[^/]+:|https://[^/]+/)([^/]+/[^/]+)`)
	subs = re.FindStringSubmatch(gitURL)
	if len(subs) != 3 {
		return "", "", errors.New("failed to get git repository name")
	}

	// USERNAME/REPOSITORY.git --> USERNAME/REPOSITORY
	repo := subs[2]
	repo = strings.TrimSuffix(repo, ".git")

	// Split repo owner and name
	subs = strings.Split(repo, "/")
	owner := subs[0]
	name := subs[1]

	return owner, name, nil
}

func (g *git) GetBranchName() (string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = g.workDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%s: %s", err.Error(), stderr.String())
	}

	branch := strings.Trim(string(stdout.String()), "\n")

	return branch, nil
}

func (g *git) GetCommitSHA(short bool) (string, error) {
	var cmd *exec.Cmd
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if short {
		cmd = exec.Command("git", "rev-parse", "--short", "HEAD")
	} else {
		cmd = exec.Command("git", "rev-parse", "HEAD")
	}

	cmd.Dir = g.workDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%s: %s", err.Error(), stderr.String())
	}

	sha := strings.Trim(string(stdout.String()), "\n")

	return sha, nil
}

func (g *git) Commit(message string, files ...string) error {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	// git add ...
	args := append([]string{"add"}, files...)
	cmd := exec.Command("git", args...)
	cmd.Dir = g.workDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %s", err.Error(), stderr.String())
	}

	stdout.Reset()
	stderr.Reset()

	// git commit -m ...
	cmd = exec.Command("git", "commit", "-m", message)
	cmd.Dir = g.workDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %s", err.Error(), stderr.String())
	}

	return nil
}

func (g *git) Tag(tag string) error {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	// git tag ...
	cmd := exec.Command("git", "tag", tag)
	cmd.Dir = g.workDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %s", err.Error(), stderr.String())
	}

	return nil
}

func (g *git) Push(includeTags bool) error {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	// git push
	cmd := exec.Command("git", "push")
	cmd.Dir = g.workDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %s", err.Error(), stderr.String())
	}

	stdout.Reset()
	stderr.Reset()

	if includeTags {
		// git push --tags
		cmd := exec.Command("git", "push", "--tags")
		cmd.Dir = g.workDir
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("%s: %s", err.Error(), stderr.String())
		}
	}

	return nil
}
