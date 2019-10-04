package action

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"

	"github.com/moorara/cherry/internal/step"
	"github.com/moorara/cherry/pkg/cui"
)

const repo = "moorara/cherry"

// update is the action for update command.
type update struct {
	ui    cui.CUI
	step1 *step.GitHubGetLatestRelease
	step2 *step.GitHubDownloadAsset
}

// NewUpdate creates an instance of Update action.
func NewUpdate(ui cui.CUI, githubToken string) Action {
	transport := &http.Transport{}
	client := &http.Client{
		Transport: transport,
	}

	return &update{
		ui: ui,
		step1: &step.GitHubGetLatestRelease{
			Client:  client,
			Token:   githubToken,
			BaseURL: step.GitHubAPIURL,
			Repo:    repo,
		},
		step2: &step.GitHubDownloadAsset{
			Client:    client,
			Token:     githubToken,
			BaseURL:   step.GitHubURL,
			Repo:      repo,
			Tag:       "TBD",
			AssetName: "TBD",
			Filepath:  "TBD",
		},
	}
}

// Dry is a dry run of the action.
func (u *update) Dry(ctx context.Context) error {
	binPath, err := exec.LookPath(os.Args[0])
	if err != nil {
		return err
	}

	// Running Dry does not set .Result.LatestRelease.TagName
	if err = u.step1.Run(ctx); err != nil {
		return err
	}

	u.step2.Tag = u.step1.Result.LatestRelease.TagName
	u.step2.AssetName = fmt.Sprintf("cherry-%s-%s", runtime.GOOS, runtime.GOARCH)
	u.step2.Filepath = binPath

	if err = u.step2.Dry(ctx); err != nil {
		return err
	}

	return nil
}

// Run executes the action.
func (u *update) Run(ctx context.Context) error {
	binPath, err := exec.LookPath(os.Args[0])
	if err != nil {
		return err
	}

	u.ui.Outputf("⬇ Getting the latest release of Cherry ...")

	if err = u.step1.Run(ctx); err != nil {
		return err
	}

	u.ui.Outputf("⬇ Downloading the latest release of Cherry ...")

	u.step2.Tag = u.step1.Result.LatestRelease.TagName
	u.step2.AssetName = fmt.Sprintf("cherry-%s-%s", runtime.GOOS, runtime.GOARCH)
	u.step2.Filepath = binPath

	if err = u.step2.Run(ctx); err != nil {
		return err
	}

	u.ui.Infof("🍒 Cherry %s installed successfully.", u.step1.Result.LatestRelease.Name)

	return nil
}

// Revert reverts back an executed action.
func (u *update) Revert(ctx context.Context) error {
	if err := u.step2.Revert(ctx); err != nil {
		return err
	}

	if err := u.step1.Revert(ctx); err != nil {
		return err
	}

	return nil
}
