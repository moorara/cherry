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

// Update is the action for update command.
type Update struct {
	ui    cui.CUI
	step1 *step.GitHubGetLatestRelease
	step2 *step.GitHubDownloadAsset
}

// NewUpdate creates an instance of Update action.
func NewUpdate(ui cui.CUI, githubToken string) *Update {
	transport := &http.Transport{}
	client := &http.Client{
		Transport: transport,
	}

	return &Update{
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
func (u *Update) Dry(ctx context.Context) error {
	binPath, err := exec.LookPath(os.Args[0])
	if err != nil {
		return err
	}

	// Running Dry does not set .Result.LatestRelease.TagName
	err = u.step1.Run(ctx)
	if err != nil {
		return err
	}

	u.step2.Tag = u.step1.Result.LatestRelease.TagName
	u.step2.AssetName = fmt.Sprintf("cherry-%s-%s", runtime.GOOS, runtime.GOARCH)
	u.step2.Filepath = binPath
	err = u.step2.Dry(ctx)
	if err != nil {
		return err
	}

	return nil
}

// Run executes the action.
func (u *Update) Run(ctx context.Context) error {
	binPath, err := exec.LookPath(os.Args[0])
	if err != nil {
		return err
	}

	u.ui.Outputf("⬇ Getting the latest release of Cherry ...")

	err = u.step1.Run(ctx)
	if err != nil {
		return err
	}

	u.ui.Outputf("⬇ Downloading the latest release of Cherry ...")

	u.step2.Tag = u.step1.Result.LatestRelease.TagName
	u.step2.AssetName = fmt.Sprintf("cherry-%s-%s", runtime.GOOS, runtime.GOARCH)
	u.step2.Filepath = binPath
	err = u.step2.Run(ctx)
	if err != nil {
		return err
	}

	u.ui.Infof("🍒 Cherry %s installed successfully.", u.step1.Result.LatestRelease.Name)

	return nil
}

// Revert reverts back an executed action.
func (u *Update) Revert(ctx context.Context) error {
	if err := u.step1.Revert(ctx); err != nil {
		return err
	}

	if err := u.step2.Revert(ctx); err != nil {
		return err
	}

	return nil
}
