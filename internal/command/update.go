package command

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/mitchellh/cli"
)

const (
	updateFlagErr   = 501
	updateGitHubErr = 502
	updateFileErr   = 503
	updateTimeout   = time.Minute

	updateSynopsis = `update cherry`
	updateHelp     = `
	Use this command for updating cherry to the latest release.

	Examples:

		cherry update
	`
)

// updateCommand implements cli.Command interface.
type updateCommand struct {
	ui cli.Ui
}

// NewUpdateCommand creates an update command.
func NewUpdateCommand(ui cli.Ui) (cli.Command, error) {
	return &updateCommand{
		ui: ui,
	}, nil
}

// Synopsis returns a short one-line synopsis of the command.
func (c *updateCommand) Synopsis() string {
	return updateSynopsis
}

// Help returns a long help text including usage, description, and list of flags for the command.
func (c *updateCommand) Help() string {
	return updateHelp
}

// Run runs the actual command with the given command-line arguments.
func (c *updateCommand) Run(args []string) int {
	fs := flag.NewFlagSet("update", flag.ContinueOnError)
	fs.Usage = func() {
		c.ui.Output(c.Help())
	}

	if err := fs.Parse(args); err != nil {
		return updateFlagErr
	}

	ctx, cancel := context.WithTimeout(context.Background(), updateTimeout)
	defer cancel()

	transport := &http.Transport{}
	client := &http.Client{
		Transport: transport,
	}

	// Run preflight checks

	var githubToken string

	{
		c.ui.Output("◉ Running preflight checks ...")

		githubToken = os.Getenv("CHERRY_GITHUB_TOKEN")
		if githubToken == "" {
			c.ui.Error("CHERRY_GITHUB_TOKEN environment variable not set.")
			return updateGitHubErr
		}
	}

	{

		url := "https://api.github.com/repos/moorara/cherry"
		req, _ := http.NewRequest("GET", url, nil)
		req = req.WithContext(ctx)
		req.Header.Set("Authorization", "token "+githubToken)
		req.Header.Set("Accept", "application/vnd.github.v3+json")
		req.Header.Set("User-Agent", "cherry") // ref: https://docs.github.com/en/rest/overview/resources-in-the-rest-api#user-agent-required
		req.Header.Set("Content-Type", "application/json")

		res, err := client.Do(req)
		if err != nil {
			c.ui.Error(fmt.Sprintf("Error on checking GitHub access: %s", err))
			return updateGitHubErr
		}
		defer res.Body.Close()

		if res.StatusCode != 200 {
			c.ui.Error(fmt.Sprintf("Error on checking GitHub access: invalid status code %d", res.StatusCode))
			return updateGitHubErr
		}
	}

	// Get the latest release of Cherry from GitHub
	// See https://docs.github.com/en/rest/reference/repos#get-the-latest-release

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
		Assets     []struct {
			ID          int    `json:"id"`
			Name        string `json:"name"`
			Label       string `json:"label"`
			State       string `json:"state"`
			Size        int    `json:"size"`
			ContentType string `json:"content_type"`
			URL         string `json:"url"`
			DownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}{}

	{
		c.ui.Output(fmt.Sprintf("⬇ Finding the latest release of Cherry ..."))

		url := "https://api.github.com/repos/moorara/cherry/releases/latest"
		req, _ := http.NewRequest("GET", url, nil)
		req = req.WithContext(ctx)
		req.Header.Set("Authorization", "token "+githubToken)
		req.Header.Set("Accept", "application/vnd.github.v3+json")
		req.Header.Set("User-Agent", "cherry") // ref: https://docs.github.com/en/rest/overview/resources-in-the-rest-api#user-agent-required
		req.Header.Set("Content-Type", "application/json")

		res, err := client.Do(req)
		if err != nil {
			c.ui.Error(fmt.Sprintf("Error on getting the latest release of Cherry from GitHub: %s", err))
			return updateGitHubErr
		}
		defer res.Body.Close()

		if res.StatusCode != 200 {
			c.ui.Error(fmt.Sprintf("Error on getting the latest release of Cherry from GitHub: invalid status code %d", res.StatusCode))
			return updateGitHubErr
		}

		err = json.NewDecoder(res.Body).Decode(&release)
		if err != nil {
			c.ui.Error(fmt.Sprintf("Error on getting the latest release of Cherry from GitHub: %s", err))
			return updateGitHubErr
		}
	}

	// Download the binary for Cherry from GitHub

	var resBody io.ReadCloser

	{
		c.ui.Output(fmt.Sprintf("⬇ Downloading Cherry %s ...", release.TagName))

		assetName := fmt.Sprintf("cherry-%s-%s", runtime.GOOS, runtime.GOARCH)
		url := fmt.Sprintf("https://github.com/moorara/cherry/releases/download/%s/%s", release.TagName, assetName)
		req, _ := http.NewRequest("GET", url, nil)
		req = req.WithContext(ctx)
		req.Header.Set("Authorization", "token "+githubToken)
		req.Header.Set("User-Agent", "cherry") // ref: https://docs.github.com/en/rest/overview/resources-in-the-rest-api#user-agent-required

		res, err := client.Do(req)
		if err != nil {
			c.ui.Error(fmt.Sprintf("Error on downloading the latest Cherry binary from GitHub: %s", err))
			return updateGitHubErr
		}
		defer res.Body.Close()

		if res.StatusCode != 200 {
			c.ui.Error(fmt.Sprintf("Error on downloading the latest Cherry binary from GitHub: invalid status code %d", res.StatusCode))
			return updateGitHubErr
		}

		resBody = res.Body
	}

	// Write the new binary to disk
	{
		binPath, err := exec.LookPath(os.Args[0])
		if err != nil {
			c.ui.Error(fmt.Sprintf("Error on getting the path for Cherry binary: %s", err))
			return updateFileErr
		}

		file, err := os.OpenFile(binPath, os.O_WRONLY, 0755)
		if err != nil {
			c.ui.Error(fmt.Sprintf("Error on openning %s for writing: %s", binPath, err))
			return updateFileErr
		}

		_, err = io.Copy(file, resBody)
		if err != nil {
			c.ui.Error(fmt.Sprintf("Error on writing to %s: %s", binPath, err))
			return updateFileErr
		}

		c.ui.Info(fmt.Sprintf("🍒 Cherry %s written to %s", release.Name, binPath))
	}

	return 0
}
