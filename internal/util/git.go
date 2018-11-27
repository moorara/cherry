package util

import (
	"errors"
	"os/exec"
	"regexp"
	"strings"
)

// IsRepoClean determines if the Git repo has any uncommitted changes
func IsRepoClean() (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	out, err := cmd.Output()
	if err != nil {
		return false, err
	}

	return len(out) == 0, nil
}

// GetRepoName returns the name of Git repo
func GetRepoName() (string, error) {
	cmd := exec.Command("git", "remote", "-v")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	// origin  git@github.com:USERNAME/REPOSITORY.git (push)     --> git@github.com:USERNAME/REPOSITORY.git
	// origin  https://github.com/USERNAME/REPOSITORY.git (push) --> https://github.com/USERNAME/REPOSITORY.git
	re := regexp.MustCompile(`origin[[:blank:]]+(.*)[[:blank:]]\(push\)`)
	subm := re.FindStringSubmatch(string(out))
	if len(subm) != 2 {
		return "", errors.New("failed to get git repository url")
	}

	gitURL := subm[1]

	// git@github.com:USERNAME/REPOSITORY.git     --> USERNAME/REPOSITORY.git
	// https://github.com/USERNAME/REPOSITORY.git --> USERNAME/REPOSITORY.git
	re = regexp.MustCompile(`(git@[^/]+:|https://[^/]+/)([^/]+/[^/]+)`)
	subm = re.FindStringSubmatch(gitURL)
	if len(subm) != 3 {
		return "", errors.New("failed to get git repository name")
	}

	repoName := subm[2]

	// USERNAME/REPOSITORY.git --> USERNAME/REPOSITORY
	repoName = strings.TrimSuffix(repoName, ".git")

	return repoName, nil
}
