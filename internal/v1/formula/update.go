package formula

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"

	"github.com/moorara/cherry/internal/service"
	"github.com/moorara/cherry/internal/util"
)

const (
	githubRepo = "moorara/cherry"
)

func (f *formula) pickRelease(releases []service.Release, assetName string) (string, string, error) {
	for _, release := range releases {
		if release.Assets != nil {
			for _, asset := range release.Assets {
				if asset.Name == assetName {
					return release.Name, asset.DownloadURL, nil
				}
			}
		}
	}

	return "", "", fmt.Errorf("cannot find a release with asset %s", assetName)
}

func (f *formula) Update(ctx context.Context, binPath string) error {
	f.Printf("⬇ Getting the list of Cherry releases ...")

	releases, err := f.github.GetReleases(ctx, githubRepo)
	if err != nil {
		return err
	}

	name := fmt.Sprintf("cherry-%s-%s", runtime.GOOS, runtime.GOARCH)
	version, url, err := f.pickRelease(releases, name)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(binPath, os.O_WRONLY, 0755)
	if err != nil {
		return err
	}

	f.Printf("⬇ Downloading the latest release of Cherry ...")

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req = req.WithContext(ctx)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return util.NewHTTPError(res)
	}

	_, err = io.Copy(file, res.Body)
	if err != nil {
		return err
	}

	f.Infof("🍒 Cherry %s installed successfully.", version)

	return nil
}
