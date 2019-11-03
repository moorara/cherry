package step

import (
	"bytes"
	"context"
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
	Mock    Step
	WorkDir string
	Result  struct {
		IsClean bool
	}
}

// Dry is a dry run of the step.
func (s *GitStatus) Dry(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Dry(ctx)
	}

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "git", "status")
	cmd.Dir = s.WorkDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("GitStatus.Dry: %s %s", err.Error(), strings.Trim(stderr.String(), "\n"))
	}

	return nil
}

// Run executes the step.
func (s *GitStatus) Run(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Run(ctx)
	}

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "git", "status", "--porcelain")
	cmd.Dir = s.WorkDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("GitStatus.Run: %s %s", err.Error(), strings.Trim(stderr.String(), "\n"))
	}

	s.Result.IsClean = len(stdout.String()) == 0

	return nil
}

// Revert reverts back an executed step.
func (s *GitStatus) Revert(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Revert(ctx)
	}

	return nil
}

// GitGetRepo runs `git remote -v` command.
type GitGetRepo struct {
	Mock    Step
	WorkDir string
	Result  struct {
		Owner string
		Name  string
		Repo  string
	}
}

// Dry is a dry run of the step.
func (s *GitGetRepo) Dry(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Dry(ctx)
	}

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "git", "remote")
	cmd.Dir = s.WorkDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("GitGetRepo.Dry: %s %s", err.Error(), strings.Trim(stderr.String(), "\n"))
	}

	return nil
}

// Run executes the step.
func (s *GitGetRepo) Run(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Run(ctx)
	}

	var err error
	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "git", "remote", "-v")
	cmd.Dir = s.WorkDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("GitGetRepo.Run: %s %s", err.Error(), strings.Trim(stderr.String(), "\n"))
	}

	owner, name, err := parseGitURL(stdout.String())
	if err != nil {
		return err
	}

	s.Result.Owner = owner
	s.Result.Name = name
	s.Result.Repo = owner + "/" + name

	return nil
}

// Revert reverts back an executed step.
func (s *GitGetRepo) Revert(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Revert(ctx)
	}

	return nil
}

// GitGetBranch runs `git rev-parse --abbrev-ref HEAD` command.
type GitGetBranch struct {
	Mock    Step
	WorkDir string
	Result  struct {
		Name string
	}
}

// Dry is a dry run of the step.
func (s *GitGetBranch) Dry(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Dry(ctx)
	}

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = s.WorkDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("GitGetBranch.Dry: %s %s", err.Error(), strings.Trim(stderr.String(), "\n"))
	}

	return nil
}

// Run executes the step.
func (s *GitGetBranch) Run(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Run(ctx)
	}

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = s.WorkDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("GitGetBranch.Run: %s %s", err.Error(), strings.Trim(stderr.String(), "\n"))
	}

	s.Result.Name = strings.Trim(stdout.String(), "\n")

	return nil
}

// Revert reverts back an executed step.
func (s *GitGetBranch) Revert(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Revert(ctx)
	}

	return nil
}

// GitGetHEAD runs `git rev-parse HEAD` command.
type GitGetHEAD struct {
	Mock    Step
	WorkDir string
	Result  struct {
		SHA      string
		ShortSHA string
	}
}

// Dry is a dry run of the step.
func (s *GitGetHEAD) Dry(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Dry(ctx)
	}

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "HEAD")
	cmd.Dir = s.WorkDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("GitGetHEAD.Dry: %s %s", err.Error(), strings.Trim(stderr.String(), "\n"))
	}

	return nil
}

// Run executes the step.
func (s *GitGetHEAD) Run(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Run(ctx)
	}

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "HEAD")
	cmd.Dir = s.WorkDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("GitGetHEAD.Run: %s %s", err.Error(), strings.Trim(stderr.String(), "\n"))
	}

	s.Result.SHA = strings.Trim(stdout.String(), "\n")
	s.Result.ShortSHA = s.Result.SHA[:7]

	return nil
}

// Revert reverts back an executed step.
func (s *GitGetHEAD) Revert(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Revert(ctx)
	}

	return nil
}

// GitAdd runs `git add <files>` command.
type GitAdd struct {
	Mock    Step
	WorkDir string
	Files   []string
}

// Dry is a dry run of the step.
func (s *GitAdd) Dry(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Dry(ctx)
	}

	var stdout, stderr bytes.Buffer
	args := append([]string{"add", "--dry-run"}, s.Files...)
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = s.WorkDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("GitAdd.Dry: %s %s", err.Error(), strings.Trim(stderr.String(), "\n"))
	}

	return nil
}

// Run executes the step.
func (s *GitAdd) Run(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Run(ctx)
	}

	var stdout, stderr bytes.Buffer
	args := append([]string{"add"}, s.Files...)
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = s.WorkDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("GitAdd.Run: %s %s", err.Error(), strings.Trim(stderr.String(), "\n"))
	}

	return nil
}

// Revert reverts back an executed step.
func (s *GitAdd) Revert(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Revert(ctx)
	}

	var stdout, stderr bytes.Buffer

	// git reset <files>
	args := append([]string{"reset"}, s.Files...)
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = s.WorkDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("GitAdd.Revert: %s %s", err.Error(), strings.Trim(stderr.String(), "\n"))
	}

	return nil
}

// GitCommit runs `git commit -m <message>` command.
type GitCommit struct {
	Mock    Step
	WorkDir string
	Message string
}

// Dry is a dry run of the step.
func (s *GitCommit) Dry(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Dry(ctx)
	}

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "git", "log", "--oneline")
	cmd.Dir = s.WorkDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("GitCommit.Dry: %s %s", err.Error(), strings.Trim(stderr.String(), "\n"))
	}

	return nil
}

// Run executes the step.
func (s *GitCommit) Run(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Run(ctx)
	}

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "git", "commit", "-m", s.Message)
	cmd.Dir = s.WorkDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("GitCommit.Run: %s %s", err.Error(), strings.Trim(stderr.String(), "\n"))
	}

	return nil
}

// Revert reverts back an executed step.
func (s *GitCommit) Revert(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Revert(ctx)
	}

	var stdout, stderr bytes.Buffer

	// git reset --soft HEAD~1
	cmd := exec.CommandContext(ctx, "git", "reset", "--soft", "HEAD~1")
	cmd.Dir = s.WorkDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("GitCommit.Revert: %s %s", err.Error(), strings.Trim(stderr.String(), "\n"))
	}

	return nil
}

// GitTag runs `git tag` or `git tag -a <tag> -m <message>` command.
type GitTag struct {
	Mock       Step
	WorkDir    string
	Tag        string
	Annotation string
}

// Dry is a dry run of the step.
func (s *GitTag) Dry(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Dry(ctx)
	}

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "git", "tag")
	cmd.Dir = s.WorkDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("GitTag.Dry: %s %s", err.Error(), strings.Trim(stderr.String(), "\n"))
	}

	return nil
}

// Run executes the step.
func (s *GitTag) Run(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Run(ctx)
	}

	var stdout, stderr bytes.Buffer

	var cmd *exec.Cmd
	if s.Annotation == "" {
		cmd = exec.CommandContext(ctx, "git", "tag", s.Tag)
	} else {
		cmd = exec.CommandContext(ctx, "git", "tag", "-a", s.Tag, "-m", s.Annotation)
	}

	cmd.Dir = s.WorkDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("GitTag.Run: %s %s", err.Error(), strings.Trim(stderr.String(), "\n"))
	}

	return nil
}

// Revert reverts back an executed step.
func (s *GitTag) Revert(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Revert(ctx)
	}

	var stdout, stderr bytes.Buffer

	// git tag --delete <tag>
	cmd := exec.CommandContext(ctx, "git", "tag", "--delete", s.Tag)
	cmd.Dir = s.WorkDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("GitTag.Revert: %s %s", err.Error(), strings.Trim(stderr.String(), "\n"))
	}

	return nil
}

// GitPush runs `git push` command.
type GitPush struct {
	Mock    Step
	WorkDir string
}

// Dry is a dry run of the step.
func (s *GitPush) Dry(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Dry(ctx)
	}

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "git", "remote", "get-url", "origin")
	cmd.Dir = s.WorkDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("GitPush.Dry: %s %s", err.Error(), strings.Trim(stderr.String(), "\n"))
	}

	return nil
}

// Run executes the step.
func (s *GitPush) Run(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Run(ctx)
	}

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "git", "push")
	cmd.Dir = s.WorkDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("GitPush.Run: %s %s", err.Error(), strings.Trim(stderr.String(), "\n"))
	}

	return nil
}

// Revert reverts back an executed step.
func (s *GitPush) Revert(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Revert(ctx)
	}

	// TODO: implement revert
	return errors.New("cannot revert git push")
}

// GitPushTag runs `git push origin <tag>` command.
type GitPushTag struct {
	Mock    Step
	WorkDir string
	Tag     string
}

// Dry is a dry run of the step.
func (s *GitPushTag) Dry(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Dry(ctx)
	}

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "git", "remote", "get-url", "origin")
	cmd.Dir = s.WorkDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("GitPushTag.Dry: %s %s", err.Error(), strings.Trim(stderr.String(), "\n"))
	}

	return nil
}

// Run executes the step.
func (s *GitPushTag) Run(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Run(ctx)
	}

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "git", "push", "origin", s.Tag)
	cmd.Dir = s.WorkDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("GitPushTag.Run: %s %s", err.Error(), strings.Trim(stderr.String(), "\n"))
	}

	return nil
}

// Revert reverts back an executed step.
func (s *GitPushTag) Revert(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Revert(ctx)
	}

	// TODO: implement revert
	return errors.New("cannot revert git push")
}

// GitPull runs `git pull` command.
type GitPull struct {
	Mock    Step
	WorkDir string
}

// Dry is a dry run of the step.
func (s *GitPull) Dry(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Dry(ctx)
	}

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "git", "remote", "get-url", "origin")
	cmd.Dir = s.WorkDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("GitPull.Dry: %s %s", err.Error(), strings.Trim(stderr.String(), "\n"))
	}

	return nil
}

// Run executes the step.
func (s *GitPull) Run(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Run(ctx)
	}

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "git", "pull")
	cmd.Dir = s.WorkDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("GitPull.Run: %s %s", err.Error(), strings.Trim(stderr.String(), "\n"))
	}

	return nil
}

// Revert reverts back an executed step.
func (s *GitPull) Revert(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Revert(ctx)
	}

	// TODO: implement revert
	return errors.New("cannot revert git pull")
}
