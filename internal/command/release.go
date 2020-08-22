package command

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	netURL "net/url"

	"github.com/mitchellh/cli"
	"github.com/moorara/cherry/internal/spec"
	"github.com/moorara/cherry/pkg/semver"
)

const (
	releaseFlagErr       = 301
	releaseOSErr         = 302
	releaseGitErr        = 303
	releaseGoErr         = 304
	releaseChangelogErr  = 305
	releaseGitHubErr     = 306
	releaseGitHubPermErr = 307
	releaseRemoteURLErr  = 308
	releaseRemoteRepoErr = 309
	releaseBranchErr     = 310
	releaseStatusErr     = 311
	releaseSemVerErr     = 312
	releaseUploadErr     = 313
	releaseTimeout       = 10 * time.Minute

	releaseSynopsis = `create a new release`
	releaseHelp     = `
	Use this command for creating a new release.
	This assumes your remote repository is named origin.
	The initial semantic version release is 0.1.0.

	Supported Remote Repositories:

		- GitHub (github.com)

	Flags:

		-patch:    create a patch version release                       (default: true)
		-minor:    create a minor version release                       (default: false)
		-major:    create a major version release                       (default: false)
		-comment:  add a comment for the release
		-build:    build the artifacts and include them in the release  (default: false)

	Examples:

		cherry release
		cherry release -build
		cherry release -minor
		cherry release -minor -build
		cherry release -major
		cherry release -major -build
		cherry release -comment "release comment"
	`
)

// release implements cli.Command interface.
type release struct {
	ui   cli.Ui
	spec spec.Spec
}

// NewRelease creates a release command.
func NewRelease(ui cli.Ui, s spec.Spec) (cli.Command, error) {
	return &release{
		ui:   ui,
		spec: s,
	}, nil
}

// Synopsis returns a short one-line synopsis of the command.
func (r *release) Synopsis() string {
	return releaseSynopsis
}

// Help returns a long help text including usage, description, and list of flags for the command.
func (r *release) Help() string {
	return releaseHelp
}

// Run runs the actual command with the given command-line arguments.
func (r *release) Run(args []string) int {
	var patch, minor, major bool
	var comment string

	fs := r.spec.Release.FlagSet()
	fs.BoolVar(&patch, "patch", true, "")
	fs.BoolVar(&minor, "minor", false, "")
	fs.BoolVar(&major, "major", false, "")
	fs.StringVar(&comment, "comment", "", "")
	fs.Usage = func() {
		r.ui.Output(r.Help())
	}

	if err := fs.Parse(args); err != nil {
		return releaseFlagErr
	}

	var version semver.Version
	switch {
	case major:
		version = semver.Major
	case minor:
		version = semver.Minor
	case patch:
		version = semver.Patch
	default:
		version = semver.Patch
	}

	ctx, cancel := context.WithTimeout(context.Background(), releaseTimeout)
	defer cancel()

	transport := &http.Transport{}
	client := &http.Client{
		Transport: transport,
	}

	// Get remote repository information

	// Run preflight checks

	var dir, githubToken string
	var repoOwner, repoName string

	{
		r.ui.Output("◉ Running preflight checks ...")

		var err error
		dir, err = os.Getwd()
		if err != nil {
			r.ui.Error(fmt.Sprintf("Error on getting the current working directory: %s", err))
			return releaseOSErr
		}

		githubToken = os.Getenv("CHERRY_GITHUB_TOKEN")
		if githubToken == "" {
			r.ui.Error("CHERRY_GITHUB_TOKEN environment variable not set.")
			return releaseGitHubErr
		}
	}

	{
		var stdout, stderr bytes.Buffer
		cmd := exec.CommandContext(ctx, "git", "version")
		cmd.Dir = dir
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			r.ui.Error(fmt.Sprintf("Error on checking git: %s %s", err, strings.Trim(stderr.String(), "\n")))
			return buildGitErr
		}

		stdout.Reset()
		stderr.Reset()
		cmd = exec.CommandContext(ctx, "go", "version")
		cmd.Dir = dir
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			r.ui.Error(fmt.Sprintf("Error on checking go: %s %s", err, strings.Trim(stderr.String(), "\n")))
			return releaseGoErr
		}

		stdout.Reset()
		stderr.Reset()
		cmd = exec.CommandContext(ctx, "github_changelog_generator", "--version")
		cmd.Dir = dir
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			r.ui.Error(fmt.Sprintf("Error on checking github_changelog_generator: %s %s", err, strings.Trim(stderr.String(), "\n")))
			return releaseChangelogErr
		}
	}

	{
		var repoDomain string
		sshRE := regexp.MustCompile(`^git@([A-Za-z][0-9A-Za-z-]+[0-9A-Za-z]\.[A-Za-z]{2,}):([A-Za-z][0-9A-Za-z-]+[0-9A-Za-z])/([A-Za-z][0-9A-Za-z-]+[0-9A-Za-z])(.git)?$`)
		httpsRE := regexp.MustCompile(`^https://([A-Za-z][0-9A-Za-z-]+[0-9A-Za-z]\.[A-Za-z]{2,})/([A-Za-z][0-9A-Za-z-]+[0-9A-Za-z])/([A-Za-z][0-9A-Za-z-]+[0-9A-Za-z])(.git)?$`)

		var stdout, stderr bytes.Buffer
		cmd := exec.CommandContext(ctx, "git", "remote", "get-url", "--push", "origin")
		cmd.Dir = dir
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			r.ui.Error(fmt.Sprintf("Error on running git remote get-url --push origin: %s %s", err, strings.Trim(stderr.String(), "\n")))
			return releaseGitErr
		}
		gitRemoteURL := strings.Trim(stdout.String(), "\n")

		if subs := sshRE.FindStringSubmatch(gitRemoteURL); len(subs) == 4 || len(subs) == 5 {
			// Git remote url is using SSH protocol
			// Example: git@github.com:moorara/cherry.git --> subs = []string{"git@github.com:moorara/cherry.git", "github.com", "moorara", "cherry", ".git"}
			repoDomain = subs[1]
			repoOwner, repoName = subs[2], subs[3]
		} else if subs := httpsRE.FindStringSubmatch(gitRemoteURL); len(subs) == 4 || len(subs) == 5 {
			// Git remote url is using HTTPS protocol
			// Example: https://github.com/moorara/cherry.git --> subs = []string{"https://github.com/moorara/cherry.git", "github.com", "moorara", "cherry", ".git"}
			repoDomain = subs[1]
			repoOwner, repoName = subs[2], subs[3]
		} else {
			r.ui.Error(fmt.Sprintf("Invalid git remote url: %s", gitRemoteURL))
			return releaseRemoteURLErr
		}

		if strings.ToLower(repoDomain) != "github.com" {
			r.ui.Error(fmt.Sprintf("Unsupported remote repository: %s", repoDomain))
			return releaseRemoteRepoErr
		}
	}

	// Check GitHub permission
	// See https://docs.github.com/en/rest/reference/repos#get-repository-permissions-for-a-user

	githubUser := struct {
		ID    int    `json:"id"`
		Login string `json:"login"`
		Name  string `json:"name"`
		Email string `json:"email"`
		URL   string `json:"url"`
	}{}

	{
		r.ui.Output("◉ Checking GitHub permission ...")

		// Get the currently authenticated user
		// See https://docs.github.com/en/rest/reference/users#get-the-authenticated-user
		url := "https://api.github.com/user"
		req, _ := http.NewRequest("GET", url, nil)
		req = req.WithContext(ctx)
		req.Header.Set("Authorization", "token "+githubToken)
		req.Header.Set("Accept", "application/vnd.github.v3+json")
		req.Header.Set("User-Agent", "cherry") // ref: https://docs.github.com/en/rest/overview/resources-in-the-rest-api#user-agent-required
		req.Header.Set("Content-Type", "application/json")

		res, err := client.Do(req)
		if err != nil {
			r.ui.Error(fmt.Sprintf("Error on getting authenticated GitHub user: %s", err))
			return releaseGitHubErr
		}
		defer res.Body.Close()

		if res.StatusCode != 200 {
			r.ui.Error(fmt.Sprintf("Error on getting authenticated GitHub user: invalid status code %d", res.StatusCode))
			return releaseGitHubErr
		}

		err = json.NewDecoder(res.Body).Decode(&githubUser)
		if err != nil {
			r.ui.Error(fmt.Sprintf("Error on getting authenticated GitHub user: %s", err))
			return releaseGitHubErr
		}

		// Check if the user has write admin access for creating release and pushing tag.
		// See https://docs.github.com/en/rest/reference/repos#get-repository-permissions-for-a-user
		url = fmt.Sprintf("https://api.github.com/repos/%s/%s/collaborators/%s/permission", repoOwner, repoName, githubUser.Login)
		req, _ = http.NewRequest("GET", url, nil)
		req = req.WithContext(ctx)
		req.Header.Set("Authorization", "token "+githubToken)
		req.Header.Set("Accept", "application/vnd.github.v3+json")
		req.Header.Set("User-Agent", "cherry") // ref: https://docs.github.com/en/rest/overview/resources-in-the-rest-api#user-agent-required
		req.Header.Set("Content-Type", "application/json")

		res, err = client.Do(req)
		if err != nil {
			r.ui.Error(fmt.Sprintf("Error on checking GitHub user permission: %s", err))
			return releaseGitHubErr
		}
		defer res.Body.Close()

		if res.StatusCode != 200 {
			r.ui.Error(fmt.Sprintf("Error on checking GitHub user permission: invalid status code %d", res.StatusCode))
			return releaseGitHubErr
		}

		githubPermission := struct {
			Permission string `json:"permission"`
		}{}

		err = json.NewDecoder(res.Body).Decode(&githubPermission)
		if err != nil {
			r.ui.Error(fmt.Sprintf("Error on checking GitHub user permission: %s", err))
			return releaseGitHubErr
		}

		if githubPermission.Permission != "admin" {
			r.ui.Error("The CHERRY_GITHUB_TOKEN does not have admin permission for releasing.")
			return releaseGitHubPermErr
		}
	}

	// Make sure the active git branch is the master branch

	var gitBranch string

	{
		var stdout, stderr bytes.Buffer
		cmd := exec.CommandContext(ctx, "git", "rev-parse", "--abbrev-ref", "HEAD")
		cmd.Dir = dir
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			r.ui.Error(fmt.Sprintf("Error on running git rev-parse --abbrev-ref HEAD: %s %s", err, strings.Trim(stderr.String(), "\n")))
			return releaseGitErr
		}
		gitBranch = strings.Trim(stdout.String(), "\n")

		if gitBranch != "master" {
			r.ui.Error("Release can only be done from master branch.")
			return releaseBranchErr
		}
	}

	// Make sure there is no uncommitted change and the current branch is clean
	{
		var stdout, stderr bytes.Buffer
		cmd := exec.CommandContext(ctx, "git", "status", "--porcelain")
		cmd.Dir = dir
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			r.ui.Error(fmt.Sprintf("Error on running git status --porcelain: %s %s", err, strings.Trim(stderr.String(), "\n")))
			return releaseGitErr
		}
		gitStatusClean := len(stdout.String()) == 0

		if !gitStatusClean {
			r.ui.Error("Working directory is not clean and has uncommitted changes.")
			return releaseStatusErr
		}
	}

	// Make sure the current branch has all the latest changes
	{
		r.ui.Output("⬇️  Pulling the latest changes on master branch ...")

		var stdout, stderr bytes.Buffer
		cmd := exec.CommandContext(ctx, "git", "pull")
		cmd.Dir = dir
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			r.ui.Error(fmt.Sprintf("Error on running git pull: %s %s", err, strings.Trim(stderr.String(), "\n")))
			return releaseGitErr
		}
	}

	// Resolve the semantic version being released

	var releaseSemVer semver.SemVer
	var releaseTag string

	{
		var stdout, stderr bytes.Buffer
		cmd := exec.CommandContext(ctx, "git", "describe", "--tags", "HEAD")
		cmd.Dir = dir
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			// 128 is returned when there is no git tag
			if exiterr, ok := err.(*exec.ExitError); !ok || exiterr.ExitCode() != 128 {
				r.ui.Error(fmt.Sprintf("Error on running git describe --tags HEAD: %s %s", err, strings.Trim(stderr.String(), "\n")))
				return releaseGitErr
			}
		}
		gitDescribe := strings.Trim(stdout.String(), "\n")

		if len(gitDescribe) == 0 {
			// No git tag found -> using the default initial semantic version for the first release
			releaseSemVer = semver.SemVer{Major: 0, Minor: 1, Patch: 0}
		} else {
			lastSemVer, ok := semver.Parse(gitDescribe)
			if !ok {
				r.ui.Error(fmt.Sprintf("Invalid git tag for semantic version: %s", gitDescribe))
				return releaseSemVerErr
			}
			releaseSemVer = lastSemVer.Next().Release(version)
		}

		releaseTag = "v" + releaseSemVer.String()
	}

	// Create a new draft GitHub release
	// See https://docs.github.com/en/rest/reference/repos#create-a-release

	release := struct {
		ID         int    `json:"id"`
		Name       string `json:"name"`
		TagName    string `json:"tag_name"`
		Target     string `json:"target_commitish"`
		Draft      bool   `json:"draft"`
		Prerelease bool   `json:"prerelease"`
		Body       string `json:"body"`
		URL        string `json:"url"`
		HTMLURL    string `json:"html_url"`
		AssetsURL  string `json:"assets_url"`
		UploadURL  string `json:"upload_url"`
	}{}

	{
		r.ui.Output(fmt.Sprintf("⬆️  Creating a draft release %s ...", ""))

		body := new(bytes.Buffer)
		_ = json.NewEncoder(body).Encode(struct {
			Name       string `json:"name"`
			TagName    string `json:"tag_name"`
			Target     string `json:"target_commitish"`
			Draft      bool   `json:"draft"`
			Prerelease bool   `json:"prerelease"`
			Body       string `json:"body"`
		}{
			Name:       releaseSemVer.String(),
			TagName:    releaseTag,
			Target:     gitBranch,
			Draft:      true,
			Prerelease: false,
		})

		url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", repoOwner, repoName)
		req, _ := http.NewRequest("POST", url, body)
		req = req.WithContext(ctx)
		req.Header.Set("Authorization", "token "+githubToken)
		req.Header.Set("Accept", "application/vnd.github.v3+json")
		req.Header.Set("User-Agent", "cherry") // ref: https://docs.github.com/en/rest/overview/resources-in-the-rest-api#user-agent-required
		req.Header.Set("Content-Type", "application/json")

		res, err := client.Do(req)
		if err != nil {
			r.ui.Error(fmt.Sprintf("Error on creating a draft GitHub release: %s", err))
			return releaseGitHubErr
		}
		defer res.Body.Close()

		if res.StatusCode != 201 {
			r.ui.Error(fmt.Sprintf("Error on creating a draft GitHub release: invalid status code %d", res.StatusCode))
			return releaseGitHubErr
		}

		err = json.NewDecoder(res.Body).Decode(&release)
		if err != nil {
			r.ui.Error(fmt.Sprintf("Error on creating a draft GitHub release: %s", err))
			return releaseGitHubErr
		}
	}

	// Generate change log

	var changelogText string
	changelogFile := "CHANGELOG.md"

	{
		r.ui.Output("➡️  Creating/Updating change log ...")

		var stdout, stderr bytes.Buffer
		cmd := exec.CommandContext(ctx,
			"github_changelog_generator",
			"--token", githubToken,
			"--user", repoOwner,
			"--project", repoName,
			"--no-filter-by-milestone",
			"--exclude-labels", "question,duplicate,invalid,wontfix",
			"--future-release", releaseTag,
		)
		cmd.Dir = dir
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			r.ui.Error(fmt.Sprintf("Error on generating change log: %s %s %s", err, strings.Trim(stdout.String(), "\n"), strings.Trim(stderr.String(), "\n")))
			return releaseChangelogErr
		}

		file, err := os.Open(filepath.Join(dir, changelogFile))
		if err != nil {
			r.ui.Error(fmt.Sprintf("Error on opening change log file: %s %s", err, strings.Trim(stderr.String(), "\n")))
			return releaseChangelogErr
		}
		defer file.Close()

		// Regex for the start of current release
		startRE, err := regexp.Compile(fmt.Sprintf(`^## \[%s\]`, releaseTag))
		if err != nil {
			r.ui.Error(fmt.Sprintf("Error on compiling regex: %s", err))
			return releaseChangelogErr
		}

		// Regex for the end of current release
		endRE, err := regexp.Compile(`^(##|\\\*)`)
		if err != nil {
			r.ui.Error(fmt.Sprintf("Error on compiling regex: %s", err))
			return releaseChangelogErr
		}

		var saveText bool
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
				changelogText += line + "\n"
			}
		}

		if err := scanner.Err(); err != nil {
			r.ui.Error(fmt.Sprintf("Error on scanning change log: %s", err))
			return releaseChangelogErr
		}

		changelogText = strings.Trim(changelogText, "\n")
	}

	// Create the release commit and tag
	{
		r.ui.Output(fmt.Sprintf("➡️  Creating release commit and tag %s ...", releaseSemVer))

		var stdout, stderr bytes.Buffer
		cmd := exec.CommandContext(ctx, "git", "add", changelogFile)
		cmd.Dir = dir
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			r.ui.Error(fmt.Sprintf("Error on running git add: %s %s", err, strings.Trim(stderr.String(), "\n")))
			return releaseGitErr
		}

		stdout.Reset()
		stderr.Reset()
		commitMessage := fmt.Sprintf("Releasing %s", releaseSemVer)
		cmd = exec.CommandContext(ctx, "git", "commit", "-m", commitMessage)
		cmd.Dir = dir
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			r.ui.Error(fmt.Sprintf("Error on running git commit -m: %s %s", err, strings.Trim(stderr.String(), "\n")))
			return releaseGitErr
		}

		stdout.Reset()
		stderr.Reset()
		annotation := fmt.Sprintf("Version %s", releaseSemVer)
		cmd = exec.CommandContext(ctx, "git", "tag", "-a", releaseTag, "-m", annotation)
		cmd.Dir = dir
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			r.ui.Error(fmt.Sprintf("Error on running git tag -a -m: %s %s", err, strings.Trim(stderr.String(), "\n")))
			return releaseGitErr
		}
	}

	// Building artifacts (binaries) and uploading them to GitHub
	// See https://developer.github.com/v3/repos/releases/#upload-a-release-asset

	if r.spec.Release.Build {
		r.ui.Output("➡️  Building artifacts ...")

		b := &build{
			ui:        r.ui,
			spec:      r.spec,
			artifacts: []string{},
		}

		code := b.Run([]string{})
		if code != 0 {
			return code
		}

		r.ui.Output(fmt.Sprintf("➡️️  Uploading artifacts to release %s ...", release.Name))

		doneCh := make(chan error, len(b.artifacts))

		for _, artifact := range b.artifacts {
			go func(artifact string) {
				assetPath := filepath.Clean(artifact)
				assetName := filepath.Base(assetPath)

				assetFile, err := os.Open(assetPath)
				if err != nil {
					doneCh <- err
					return
				}
				defer assetFile.Close()

				stat, err := assetFile.Stat()
				if err != nil {
					doneCh <- err
					return
				}

				// Read the first 512 bytes of file to determine the mime type of asset
				buff := make([]byte, 512)
				_, err = assetFile.Read(buff)
				if err != nil {
					doneCh <- err
					return
				}

				// Determine mime type of asset
				// http.DetectContentType will return "application/octet-stream" if it cannot determine a more specific one
				mimeType := http.DetectContentType(buff)

				// Reset the offset back to the beginning of the file
				_, err = assetFile.Seek(0, io.SeekStart)
				if err != nil {
					doneCh <- err
					return
				}

				re := regexp.MustCompile(`\{\?[0-9A-Za-z_,]+\}`)
				url := re.ReplaceAllLiteralString(release.UploadURL, "")
				url = fmt.Sprintf("%s?name=%s", url, netURL.QueryEscape(assetName))
				req, _ := http.NewRequest("POST", url, assetFile)
				req = req.WithContext(ctx)
				req.Header.Set("Authorization", "token "+githubToken)
				req.Header.Set("Accept", "application/vnd.github.v3+json")
				req.Header.Set("User-Agent", "cherry") // ref: https://docs.github.com/en/rest/overview/resources-in-the-rest-api#user-agent-required
				req.Header.Set("Content-Type", mimeType)
				req.ContentLength = stat.Size()

				res, err := client.Do(req)
				if err != nil {
					doneCh <- err
					return
				}
				defer res.Body.Close()

				if res.StatusCode != 201 {
					doneCh <- fmt.Errorf("invalid status code %d", res.StatusCode)
					return
				}

				asset := struct {
					ID          int    `json:"id"`
					Name        string `json:"name"`
					Label       string `json:"label"`
					State       string `json:"state"`
					Size        int    `json:"size"`
					ContentType string `json:"content_type"`
					URL         string `json:"url"`
					DownloadURL string `json:"browser_download_url"`
				}{}

				err = json.NewDecoder(res.Body).Decode(&asset)
				if err != nil {
					doneCh <- err
					return
				}

				doneCh <- nil
			}(artifact)
		}

		for range b.artifacts {
			if err := <-doneCh; err != nil {
				r.ui.Error(fmt.Sprintf("Error on uploading artifact: %s", err))
				return releaseUploadErr
			}
		}
	}

	// Enable direct push to master and defering disabling it back

	{
		r.ui.Warn("🔓 Temporarily enabling push to master branch ...")

		url := fmt.Sprintf("https://api.github.com/repos/%s/%s/branches/%s/protection/enforce_admins", repoOwner, repoName, gitBranch)
		req, _ := http.NewRequest("DELETE", url, nil)
		req = req.WithContext(ctx)
		req.Header.Set("Authorization", "token "+githubToken)
		req.Header.Set("Accept", "application/vnd.github.v3+json")
		req.Header.Set("User-Agent", "cherry") // ref: https://docs.github.com/en/rest/overview/resources-in-the-rest-api#user-agent-required
		req.Header.Set("Content-Type", "application/json")

		res, err := client.Do(req)
		if err != nil {
			r.ui.Error(fmt.Sprintf("Error on disabling push to master: %s", err))
			return releaseGitHubErr
		}
		defer res.Body.Close()

		if res.StatusCode != 204 {
			r.ui.Error(fmt.Sprintf("Error on disabling push to master: invalid status code %d", res.StatusCode))
			return releaseGitHubErr
		}

		// Make sure we re-enable the master branch protection
		defer func() {
			r.ui.Warn("🔒 Re-disabling push to master branch ...")

			url := fmt.Sprintf("https://api.github.com/repos/%s/%s/branches/%s/protection/enforce_admins", repoOwner, repoName, gitBranch)
			req, _ := http.NewRequest("POST", url, nil)
			req = req.WithContext(ctx)
			req.Header.Set("Authorization", "token "+githubToken)
			req.Header.Set("Accept", "application/vnd.github.v3+json")
			req.Header.Set("User-Agent", "cherry") // ref: https://docs.github.com/en/rest/overview/resources-in-the-rest-api#user-agent-required
			req.Header.Set("Content-Type", "application/json")

			res, err := client.Do(req)
			if err != nil {
				r.ui.Error(fmt.Sprintf("Error on enabling push to master: %s", err))
			}
			defer res.Body.Close()

			if res.StatusCode != 200 {
				r.ui.Error(fmt.Sprintf("Error on enabling push to master: invalid status code %d", res.StatusCode))
			}
		}()
	}

	// Push release commit to GitHub
	{
		r.ui.Info(fmt.Sprintf("⬆️  Pushing release commit %s ...", releaseSemVer))

		var stdout, stderr bytes.Buffer
		cmd := exec.CommandContext(ctx, "git", "push")
		cmd.Dir = dir
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			r.ui.Error(fmt.Sprintf("Error on running git push: %s %s", err, strings.Trim(stderr.String(), "\n")))
			return releaseGitErr
		}
	}

	// Push release tag to GitHub
	{
		r.ui.Info(fmt.Sprintf("⬆️  Pushing release tag %s ...", releaseTag))

		var stdout, stderr bytes.Buffer
		cmd := exec.CommandContext(ctx, "git", "push", "origin", releaseTag)
		cmd.Dir = dir
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			r.ui.Error(fmt.Sprintf("Error on running git push origin: %s %s", err, strings.Trim(stderr.String(), "\n")))
			return releaseGitErr
		}
	}

	// Publishing GitHub release
	{
		r.ui.Info(fmt.Sprintf("⬆️  Publishing release %s ...", release.Name))

		body := new(bytes.Buffer)
		_ = json.NewEncoder(body).Encode(struct {
			Name       string `json:"name"`
			TagName    string `json:"tag_name"`
			Target     string `json:"target_commitish"`
			Draft      bool   `json:"draft"`
			Prerelease bool   `json:"prerelease"`
			Body       string `json:"body"`
		}{
			Name:       releaseSemVer.String(),
			TagName:    releaseTag,
			Target:     gitBranch,
			Draft:      false,
			Prerelease: false,
			Body:       fmt.Sprintf("%s\n\n%s", comment, changelogText),
		})

		url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/%d", repoOwner, repoName, release.ID)
		req, _ := http.NewRequest("PATCH", url, body)
		req = req.WithContext(ctx)
		req.Header.Set("Authorization", "token "+githubToken)
		req.Header.Set("Accept", "application/vnd.github.v3+json")
		req.Header.Set("User-Agent", "cherry") // ref: https://docs.github.com/en/rest/overview/resources-in-the-rest-api#user-agent-required
		req.Header.Set("Content-Type", "application/json")

		res, err := client.Do(req)
		if err != nil {
			r.ui.Error(fmt.Sprintf("Error on editing GitHub release: %s", err))
			return releaseGitHubErr
		}
		defer res.Body.Close()

		if res.StatusCode != 200 {
			r.ui.Error(fmt.Sprintf("Error on editing GitHub release: invalid status code %d", res.StatusCode))
			return releaseGitHubErr
		}

		err = json.NewDecoder(res.Body).Decode(&release)
		if err != nil {
			r.ui.Error(fmt.Sprintf("Error on editing GitHub release: %s", err))
			return releaseGitHubErr
		}
	}

	return 0
}
