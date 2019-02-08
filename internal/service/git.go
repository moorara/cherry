package service

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
		GetRepo() (*Repo, error)
		GetBranch() (*Branch, error)
		GetHEAD() (*Commit, error)
		Commit(message string, files ...string) error
		Tag(tag string) error
		Push() error
		PushTag(tag string) error
		Pull() error
	}

	git struct {
		workDir string
	}

	// Repo is the model for a git repository
	Repo struct {
		Owner string
		Name  string
	}

	// Branch is the model for a git branch
	Branch struct {
		Name string
	}

	// Commit is the model for a git commit
	Commit struct {
		SHA      string
		ShortSHA string
	}
)

// NewGit creates a new git client
func NewGit(workDir string) Git {
	return &git{
		workDir: workDir,
	}
}

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

func (g *git) GetRepo() (*Repo, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := exec.Command("git", "remote", "-v")
	cmd.Dir = g.workDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("%s: %s", err.Error(), stderr.String())
	}

	// origin  git@github.com:USERNAME/REPOSITORY.git (push)     --> git@github.com:USERNAME/REPOSITORY.git
	// origin  https://github.com/USERNAME/REPOSITORY.git (push) --> https://github.com/USERNAME/REPOSITORY.git
	re := regexp.MustCompile(`origin[[:blank:]]+(.*)[[:blank:]]\(push\)`)
	subs := re.FindStringSubmatch(string(stdout.String()))
	if len(subs) != 2 {
		return nil, errors.New("failed to get git repository url")
	}

	gitURL := subs[1]

	// git@github.com:USERNAME/REPOSITORY.git     --> USERNAME/REPOSITORY.git
	// https://github.com/USERNAME/REPOSITORY.git --> USERNAME/REPOSITORY.git
	re = regexp.MustCompile(`(git@[^/]+:|https://[^/]+/)([^/]+/[^/]+)`)
	subs = re.FindStringSubmatch(gitURL)
	if len(subs) != 3 {
		return nil, errors.New("failed to get git repository name")
	}

	// USERNAME/REPOSITORY.git --> USERNAME/REPOSITORY
	repo := subs[2]
	repo = strings.TrimSuffix(repo, ".git")

	// Split repo owner and name
	subs = strings.Split(repo, "/")
	owner := subs[0]
	name := subs[1]

	return &Repo{
		Owner: owner,
		Name:  name,
	}, nil
}

func (g *git) GetBranch() (*Branch, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = g.workDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("%s: %s", err.Error(), stderr.String())
	}

	name := strings.Trim(string(stdout.String()), "\n")

	return &Branch{
		Name: name,
	}, nil
}

func (g *git) GetHEAD() (*Commit, error) {
	var cmd *exec.Cmd
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd = exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = g.workDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("%s: %s", err.Error(), stderr.String())
	}

	sha := strings.Trim(string(stdout.String()), "\n")
	short := sha[:7]

	return &Commit{
		SHA:      sha,
		ShortSHA: short,
	}, nil
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

func (g *git) Push() error {
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

	return nil
}

func (g *git) PushTag(tag string) error {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	// git push
	cmd := exec.Command("git", "push", "origin", tag)
	cmd.Dir = g.workDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %s", err.Error(), stderr.String())
	}

	return nil
}

func (g *git) Pull() error {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	// git tag ...
	cmd := exec.Command("git", "pull")
	cmd.Dir = g.workDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %s", err.Error(), stderr.String())
	}

	return nil
}

// Path returns the owner/name combination
func (r *Repo) Path() string {
	return r.Owner + "/" + r.Name
}
