package git

import (
	"errors"
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
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = g.workDir
	out, err := cmd.Output()
	if err != nil {
		return false, err
	}

	return len(out) == 0, nil
}

// GetRepoName returns the owner and name of Git repo
func (g *git) GetRepoName() (string, string, error) {
	cmd := exec.Command("git", "remote", "-v")
	cmd.Dir = g.workDir
	out, err := cmd.Output()
	if err != nil {
		return "", "", err
	}

	// origin  git@github.com:USERNAME/REPOSITORY.git (push)     --> git@github.com:USERNAME/REPOSITORY.git
	// origin  https://github.com/USERNAME/REPOSITORY.git (push) --> https://github.com/USERNAME/REPOSITORY.git
	re := regexp.MustCompile(`origin[[:blank:]]+(.*)[[:blank:]]\(push\)`)
	subs := re.FindStringSubmatch(string(out))
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
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = g.workDir
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	branch := strings.Trim(string(out), "\n")

	return branch, nil
}

func (g *git) GetCommitSHA(short bool) (string, error) {
	var cmd *exec.Cmd

	if short {
		cmd = exec.Command("git", "rev-parse", "--short", "HEAD")
	} else {
		cmd = exec.Command("git", "rev-parse", "HEAD")
	}

	cmd.Dir = g.workDir
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	sha := strings.Trim(string(out), "\n")

	return sha, nil
}

func (g *git) Commit(message string, files ...string) error {
	// git add ...
	args := append([]string{"add"}, files...)
	cmd := exec.Command("git", args...)
	cmd.Dir = g.workDir
	if err := cmd.Run(); err != nil {
		return err
	}

	// git commit -m ...
	cmd = exec.Command("git", "commit", "-m", message)
	cmd.Dir = g.workDir
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func (g *git) Tag(tag string) error {
	// git tag ...
	cmd := exec.Command("git", "tag", tag)
	cmd.Dir = g.workDir
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func (g *git) Push(includeTags bool) error {
	// git push
	cmd := exec.Command("git", "push")
	cmd.Dir = g.workDir
	if err := cmd.Run(); err != nil {
		return err
	}

	if includeTags {
		// git push --tags
		cmd := exec.Command("git", "push", "--tags")
		cmd.Dir = g.workDir
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	return nil
}
