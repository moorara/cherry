package step

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

func parseGitURL(output string) (string, string, error) {
	// origin  git@github.com:USERNAME/REPOSITORY.git (push)     --> git@github.com:USERNAME/REPOSITORY.git
	// origin  https://github.com/USERNAME/REPOSITORY.git (push) --> https://github.com/USERNAME/REPOSITORY.git
	re := regexp.MustCompile(`origin[[:blank:]]+(.*)[[:blank:]]\(push\)`)
	subs := re.FindStringSubmatch(output)
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

// GitStatus runs `git status --porcelain` command.
type GitStatus struct {
	WorkDir string
	Result  struct {
		IsClean bool
	}
}

// Dry is a dry run of the step.
func (s *GitStatus) Dry() error {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("git", "status")
	cmd.Dir = s.WorkDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %s", err.Error(), strings.Trim(stderr.String(), "\n"))
	}

	return nil
}

// Run executes the step.
func (s *GitStatus) Run() error {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = s.WorkDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %s", err.Error(), strings.Trim(stderr.String(), "\n"))
	}

	s.Result.IsClean = len(stdout.String()) == 0

	return nil
}

// Revert reverts back an executed step.
func (s *GitStatus) Revert() error {
	return nil
}

// GitGetRepo runs `git remote -v` command.
type GitGetRepo struct {
	WorkDir string
	Result  struct {
		Owner string
		Name  string
	}
}

// Dry is a dry run of the step.
func (s *GitGetRepo) Dry() error {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("git", "remote")
	cmd.Dir = s.WorkDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %s", err.Error(), strings.Trim(stderr.String(), "\n"))
	}

	return nil
}

// Run executes the step.
func (s *GitGetRepo) Run() error {
	var err error
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("git", "remote", "-v")
	cmd.Dir = s.WorkDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %s", err.Error(), strings.Trim(stderr.String(), "\n"))
	}

	s.Result.Owner, s.Result.Name, err = parseGitURL(stdout.String())
	if err != nil {
		return err
	}

	return nil
}

// Revert reverts back an executed step.
func (s *GitGetRepo) Revert() error {
	return nil
}

// GitGetBranch runs `git rev-parse --abbrev-ref HEAD` command.
type GitGetBranch struct {
	WorkDir string
	Result  struct {
		Name string
	}
}

// Dry is a dry run of the step.
func (s *GitGetBranch) Dry() error {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = s.WorkDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %s", err.Error(), strings.Trim(stderr.String(), "\n"))
	}

	return nil
}

// Run executes the step.
func (s *GitGetBranch) Run() error {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = s.WorkDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %s", err.Error(), strings.Trim(stderr.String(), "\n"))
	}

	s.Result.Name = strings.Trim(stdout.String(), "\n")

	return nil
}

// Revert reverts back an executed step.
func (s *GitGetBranch) Revert() error {
	return nil
}

// GitGetHEAD runs `git rev-parse HEAD` command.
type GitGetHEAD struct {
	WorkDir string
	Result  struct {
		SHA      string
		ShortSHA string
	}
}

// Dry is a dry run of the step.
func (s *GitGetHEAD) Dry() error {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = s.WorkDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %s", err.Error(), strings.Trim(stderr.String(), "\n"))
	}

	return nil
}

// Run executes the step.
func (s *GitGetHEAD) Run() error {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = s.WorkDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %s", err.Error(), strings.Trim(stderr.String(), "\n"))
	}

	s.Result.SHA = strings.Trim(stdout.String(), "\n")
	s.Result.ShortSHA = s.Result.SHA[:7]

	return nil
}

// Revert reverts back an executed step.
func (s *GitGetHEAD) Revert() error {
	return nil
}

// GitAdd runs `git add <files>` command.
type GitAdd struct {
	WorkDir string
	Files   []string
}

// Dry is a dry run of the step.
func (s *GitAdd) Dry() error {
	var stdout, stderr bytes.Buffer
	args := append([]string{"add", "--dry-run"}, s.Files...)
	cmd := exec.Command("git", args...)
	cmd.Dir = s.WorkDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %s", err.Error(), strings.Trim(stderr.String(), "\n"))
	}

	return nil
}

// Run executes the step.
func (s *GitAdd) Run() error {
	var stdout, stderr bytes.Buffer
	args := append([]string{"add"}, s.Files...)
	cmd := exec.Command("git", args...)
	cmd.Dir = s.WorkDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %s", err.Error(), strings.Trim(stderr.String(), "\n"))
	}

	return nil
}

// Revert reverts back an executed step.
func (s *GitAdd) Revert() error {
	var stdout, stderr bytes.Buffer

	// git reset <files>
	args := append([]string{"reset"}, s.Files...)
	cmd := exec.Command("git", args...)
	cmd.Dir = s.WorkDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %s", err.Error(), strings.Trim(stderr.String(), "\n"))
	}

	return nil
}

// GitCommit runs `git commit -m <message>` command.
type GitCommit struct {
	WorkDir string
	Message string
}

// Dry is a dry run of the step.
func (s *GitCommit) Dry() error {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("git", "commit", "--dry-run", "-m", s.Message)
	cmd.Dir = s.WorkDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %s", err.Error(), strings.Trim(stderr.String(), "\n"))
	}

	return nil
}

// Run executes the step.
func (s *GitCommit) Run() error {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("git", "commit", "-m", s.Message)
	cmd.Dir = s.WorkDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %s", err.Error(), strings.Trim(stderr.String(), "\n"))
	}

	return nil
}

// Revert reverts back an executed step.
func (s *GitCommit) Revert() error {
	var stdout, stderr bytes.Buffer

	// git reset --soft HEAD~1
	cmd := exec.Command("git", "reset", "--soft", "HEAD~1")
	cmd.Dir = s.WorkDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %s", err.Error(), strings.Trim(stderr.String(), "\n"))
	}

	return nil
}

// GitTag runs `git tag` or `git tag -a <tag> -m <message>` command.
type GitTag struct {
	WorkDir    string
	Tag        string
	Annotation string
}

// Dry is a dry run of the step.
func (s *GitTag) Dry() error {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("git", "tag")
	cmd.Dir = s.WorkDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %s", err.Error(), strings.Trim(stderr.String(), "\n"))
	}

	return nil
}

// Run executes the step.
func (s *GitTag) Run() error {
	var stdout, stderr bytes.Buffer

	var cmd *exec.Cmd
	if s.Annotation == "" {
		cmd = exec.Command("git", "tag", s.Tag)
	} else {
		cmd = exec.Command("git", "tag", "-a", s.Tag, "-m", s.Annotation)
	}

	cmd.Dir = s.WorkDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %s", err.Error(), strings.Trim(stderr.String(), "\n"))
	}

	return nil
}

// Revert reverts back an executed step.
func (s *GitTag) Revert() error {
	var stdout, stderr bytes.Buffer

	// git tag --delete <tag>
	cmd := exec.Command("git", "tag", "--delete", s.Tag)
	cmd.Dir = s.WorkDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %s", err.Error(), strings.Trim(stderr.String(), "\n"))
	}

	return nil
}

// GitPush runs `git push` command.
type GitPush struct {
	WorkDir string
}

// Dry is a dry run of the step.
func (s *GitPush) Dry() error {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = s.WorkDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %s", err.Error(), strings.Trim(stderr.String(), "\n"))
	}

	return nil
}

// Run executes the step.
func (s *GitPush) Run() error {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("git", "push")
	cmd.Dir = s.WorkDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %s", err.Error(), strings.Trim(stderr.String(), "\n"))
	}

	return nil
}

// Revert reverts back an executed step.
func (s *GitPush) Revert() error {
	// TODO: implement revert
	return errors.New("cannot revert git push")
}

// GitPushTag runs `git push origin <tag>` command.
type GitPushTag struct {
	WorkDir string
	Tag     string
}

// Dry is a dry run of the step.
func (s *GitPushTag) Dry() error {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = s.WorkDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %s", err.Error(), strings.Trim(stderr.String(), "\n"))
	}

	return nil
}

// Run executes the step.
func (s *GitPushTag) Run() error {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("git", "push", "origin", s.Tag)
	cmd.Dir = s.WorkDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %s", err.Error(), strings.Trim(stderr.String(), "\n"))
	}

	return nil
}

// Revert reverts back an executed step.
func (s *GitPushTag) Revert() error {
	// TODO: implement revert
	return errors.New("cannot revert git push")
}

// GitPull runs `git pull` command.
type GitPull struct {
	WorkDir string
}

// Dry is a dry run of the step.
func (s *GitPull) Dry() error {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = s.WorkDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %s", err.Error(), strings.Trim(stderr.String(), "\n"))
	}

	return nil
}

// Run executes the step.
func (s *GitPull) Run() error {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("git", "pull")
	cmd.Dir = s.WorkDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %s", err.Error(), strings.Trim(stderr.String(), "\n"))
	}

	return nil
}

// Revert reverts back an executed step.
func (s *GitPull) Revert() error {
	// TODO: implement revert
	return errors.New("cannot revert git pull")
}
