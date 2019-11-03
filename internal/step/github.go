package step

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	netURL "net/url"
)

const (
	// GitHubURL is the BaseURL for GitHub.
	GitHubURL = "https://github.com"

	// GitHubAPIURL is the BaseURL for GitHub API.
	GitHubAPIURL = "https://api.github.com"
)

type (
	// GitHubReleaseData is used for creating or modifying a release.
	GitHubReleaseData struct {
		Name       string `json:"name"`
		TagName    string `json:"tag_name"`
		Target     string `json:"target_commitish"`
		Draft      bool   `json:"draft"`
		Prerelease bool   `json:"prerelease"`
		Body       string `json:"body"`
	}

	// GitHubRelease represents a GitHub release.
	GitHubRelease struct {
		ID         int           `json:"id"`
		Name       string        `json:"name"`
		TagName    string        `json:"tag_name"`
		Target     string        `json:"target_commitish"`
		Draft      bool          `json:"draft"`
		Prerelease bool          `json:"prerelease"`
		Body       string        `json:"body"`
		URL        string        `json:"url"`
		HTMLURL    string        `json:"html_url"`
		AssetsURL  string        `json:"assets_url"`
		UploadURL  string        `json:"upload_url"`
		Assets     []GitHubAsset `json:"assets"`
	}

	// GitHubAsset represents an asset for a GitHub release.
	GitHubAsset struct {
		ID          int    `json:"id"`
		Name        string `json:"name"`
		Label       string `json:"label"`
		State       string `json:"state"`
		Size        int    `json:"size"`
		ContentType string `json:"content_type"`
		URL         string `json:"url"`
		DownloadURL string `json:"browser_download_url"`
	}
)

// httpError is an http error.
type httpError struct {
	Request    *http.Request
	StatusCode int
	Message    string
}

// newHTTPError creates a new instance of httpError.
func newHTTPError(res *http.Response) *httpError {
	err := &httpError{
		Request:    res.Request,
		StatusCode: res.StatusCode,
	}

	if res.Body != nil {
		if data, e := ioutil.ReadAll(res.Body); e == nil {
			err.Message = string(data)
		}
	}

	return err
}

func (e *httpError) Error() string {
	return fmt.Sprintf("%s %s %d: %s", e.Request.Method, e.Request.URL.Path, e.StatusCode, e.Message)
}

func createGitHubRequest(ctx context.Context, token, method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)
	req.Header.Set("Authorization", fmt.Sprintf("token %s", token))
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "cherry") // ref: https://developer.github.com/v3/#user-agent-required
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// GitHubBranchProtection enables/disables branch protection for administrators.
// See https://developer.github.com/v3/repos/branches/#get-admin-enforcement-of-protected-branch
// See https://developer.github.com/v3/repos/branches/#add-admin-enforcement-of-protected-branch
// See https://developer.github.com/v3/repos/branches/#remove-admin-enforcement-of-protected-branch
type GitHubBranchProtection struct {
	Mock    Step
	Client  *http.Client
	Token   string
	BaseURL string
	Repo    string
	Branch  string
	Enabled bool
}

// Dry is a dry run of the step.
func (s *GitHubBranchProtection) Dry(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Dry(ctx)
	}

	url := fmt.Sprintf("%s/repos/%s/branches/%s/protection/enforce_admins", s.BaseURL, s.Repo, s.Branch)
	req, err := createGitHubRequest(ctx, s.Token, "GET", url, nil)
	if err != nil {
		return err
	}

	res, err := s.Client.Do(req)
	if err != nil {
		return fmt.Errorf("GitHubBranchProtection.Dry: %s", err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("GitHubBranchProtection.Dry: %s", newHTTPError(res))
	}

	return nil
}

// Run executes the step.
func (s *GitHubBranchProtection) Run(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Run(ctx)
	}

	var method string
	var statusCode int

	if s.Enabled {
		method = "POST"
		statusCode = 200
	} else {
		method = "DELETE"
		statusCode = 204
	}

	url := fmt.Sprintf("%s/repos/%s/branches/%s/protection/enforce_admins", s.BaseURL, s.Repo, s.Branch)
	req, err := createGitHubRequest(ctx, s.Token, method, url, nil)
	if err != nil {
		return err
	}

	res, err := s.Client.Do(req)
	if err != nil {
		return fmt.Errorf("GitHubBranchProtection.Run: %s", err)
	}
	defer res.Body.Close()

	if res.StatusCode != statusCode {
		return fmt.Errorf("GitHubBranchProtection.Run: %s", newHTTPError(res))
	}

	return nil
}

// Revert reverts back an executed step.
func (s *GitHubBranchProtection) Revert(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Revert(ctx)
	}

	var method string
	var statusCode int

	if s.Enabled {
		method = "DELETE"
		statusCode = 204
	} else {
		method = "POST"
		statusCode = 200
	}

	url := fmt.Sprintf("%s/repos/%s/branches/%s/protection/enforce_admins", s.BaseURL, s.Repo, s.Branch)
	req, err := createGitHubRequest(ctx, s.Token, method, url, nil)
	if err != nil {
		return err
	}

	res, err := s.Client.Do(req)
	if err != nil {
		return fmt.Errorf("GitHubBranchProtection.Revert: %s", err)
	}
	defer res.Body.Close()

	if res.StatusCode != statusCode {
		return fmt.Errorf("GitHubBranchProtection.Revert: %s", newHTTPError(res))
	}

	return nil
}

// GitHubGetLatestRelease gets the latest release.
// See https://developer.github.com/v3/repos/releases/#get-the-latest-release
type GitHubGetLatestRelease struct {
	Mock    Step
	Client  *http.Client
	Token   string
	BaseURL string
	Repo    string
	Result  struct {
		LatestRelease GitHubRelease
	}
}

// Dry is a dry run of the step.
func (s *GitHubGetLatestRelease) Dry(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Dry(ctx)
	}

	url := fmt.Sprintf("%s/repos/%s/releases/latest", s.BaseURL, s.Repo)
	req, err := createGitHubRequest(ctx, s.Token, "GET", url, nil)
	if err != nil {
		return err
	}

	res, err := s.Client.Do(req)
	if err != nil {
		return fmt.Errorf("GitHubGetLatestRelease.Dry: %s", err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("GitHubGetLatestRelease.Dry: %s", newHTTPError(res))
	}

	return nil
}

// Run executes the step.
func (s *GitHubGetLatestRelease) Run(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Run(ctx)
	}

	url := fmt.Sprintf("%s/repos/%s/releases/latest", s.BaseURL, s.Repo)
	req, err := createGitHubRequest(ctx, s.Token, "GET", url, nil)
	if err != nil {
		return err
	}

	res, err := s.Client.Do(req)
	if err != nil {
		return fmt.Errorf("GitHubGetLatestRelease.Run: %s", err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("GitHubGetLatestRelease.Run: %s", newHTTPError(res))
	}

	err = json.NewDecoder(res.Body).Decode(&s.Result.LatestRelease)
	if err != nil {
		return err
	}

	return nil
}

// Revert reverts back an executed step.
func (s *GitHubGetLatestRelease) Revert(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Revert(ctx)
	}

	return nil
}

// GitHubCreateRelease creates a new GitHub release.
// See https://developer.github.com/v3/repos/releases/#get-the-latest-release
// See https://developer.github.com/v3/repos/releases/#create-a-release
// See https://developer.github.com/v3/repos/releases/#delete-a-release
type GitHubCreateRelease struct {
	Mock        Step
	Client      *http.Client
	Token       string
	BaseURL     string
	Repo        string
	ReleaseData GitHubReleaseData
	Result      struct {
		Release GitHubRelease
	}
}

// Dry is a dry run of the step.
func (s *GitHubCreateRelease) Dry(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Dry(ctx)
	}

	url := fmt.Sprintf("%s/repos/%s/releases", s.BaseURL, s.Repo)
	req, err := createGitHubRequest(ctx, s.Token, "GET", url, nil)
	if err != nil {
		return err
	}

	res, err := s.Client.Do(req)
	if err != nil {
		return fmt.Errorf("GitHubCreateRelease.Dry: %s", err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("GitHubCreateRelease.Dry: %s", newHTTPError(res))
	}

	return nil
}

// Run executes the step.
func (s *GitHubCreateRelease) Run(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Run(ctx)
	}

	url := fmt.Sprintf("%s/repos/%s/releases", s.BaseURL, s.Repo)
	body := new(bytes.Buffer)
	_ = json.NewEncoder(body).Encode(s.ReleaseData)
	req, err := createGitHubRequest(ctx, s.Token, "POST", url, body)
	if err != nil {
		return err
	}

	res, err := s.Client.Do(req)
	if err != nil {
		return fmt.Errorf("GitHubCreateRelease.Run: %s", err)
	}
	defer res.Body.Close()

	if res.StatusCode != 201 {
		return fmt.Errorf("GitHubCreateRelease.Run: %s", newHTTPError(res))
	}

	err = json.NewDecoder(res.Body).Decode(&s.Result.Release)
	if err != nil {
		return err
	}

	return nil
}

// Revert reverts back an executed step.
func (s *GitHubCreateRelease) Revert(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Revert(ctx)
	}

	url := fmt.Sprintf("%s/repos/%s/releases/%d", s.BaseURL, s.Repo, s.Result.Release.ID)
	req, err := createGitHubRequest(ctx, s.Token, "DELETE", url, nil)
	if err != nil {
		return err
	}

	res, err := s.Client.Do(req)
	if err != nil {
		return fmt.Errorf("GitHubCreateRelease.Revert: %s", err)
	}
	defer res.Body.Close()

	if res.StatusCode != 204 {
		return fmt.Errorf("GitHubCreateRelease.Revert: %s", newHTTPError(res))
	}

	return nil
}

// GitHubEditRelease edits an existing GitHub release.
// See https://developer.github.com/v3/repos/releases/#get-a-single-release
// See https://developer.github.com/v3/repos/releases/#edit-a-release
type GitHubEditRelease struct {
	Mock        Step
	Client      *http.Client
	Token       string
	BaseURL     string
	Repo        string
	ReleaseID   int
	ReleaseData GitHubReleaseData
	Result      struct {
		Release GitHubRelease
	}
}

// Dry is a dry run of the step.
func (s *GitHubEditRelease) Dry(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Dry(ctx)
	}

	url := fmt.Sprintf("%s/repos/%s/releases", s.BaseURL, s.Repo)
	req, err := createGitHubRequest(ctx, s.Token, "GET", url, nil)
	if err != nil {
		return err
	}

	res, err := s.Client.Do(req)
	if err != nil {
		return fmt.Errorf("GitHubEditRelease.Dry: %s", err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("GitHubEditRelease.Dry: %s", newHTTPError(res))
	}

	return nil
}

// Run executes the step.
func (s *GitHubEditRelease) Run(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Run(ctx)
	}

	url := fmt.Sprintf("%s/repos/%s/releases/%d", s.BaseURL, s.Repo, s.ReleaseID)
	body := new(bytes.Buffer)
	_ = json.NewEncoder(body).Encode(s.ReleaseData)
	req, err := createGitHubRequest(ctx, s.Token, "PATCH", url, body)
	if err != nil {
		return err
	}

	res, err := s.Client.Do(req)
	if err != nil {
		return fmt.Errorf("GitHubEditRelease.Run: %s", err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("GitHubEditRelease.Run: %s", newHTTPError(res))
	}

	err = json.NewDecoder(res.Body).Decode(&s.Result.Release)
	if err != nil {
		return err
	}

	return nil
}

// Revert reverts back an executed step.
func (s *GitHubEditRelease) Revert(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Revert(ctx)
	}

	// TODO: how to revert an edited release?
	return nil
}

// GitHubUploadAssets uploads assets (files) to GitHub for a release.
// See https://developer.github.com/v3/repos/releases/#list-assets-for-a-release
// See https://developer.github.com/v3/repos/releases/#upload-a-release-asset
// See https://developer.github.com/v3/repos/releases/#delete-a-release-asset
type GitHubUploadAssets struct {
	Mock             Step
	Client           *http.Client
	Token            string
	BaseURL          string
	Repo             string
	ReleaseID        int
	ReleaseUploadURL string
	AssetFiles       []string
	Result           struct {
		Assets []GitHubAsset
	}
}

// Dry is a dry run of the step
func (s *GitHubUploadAssets) Dry(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Dry(ctx)
	}

	url := fmt.Sprintf("%s/repos/%s/releases", s.BaseURL, s.Repo)
	req, err := createGitHubRequest(ctx, s.Token, "GET", url, nil)
	if err != nil {
		return err
	}

	res, err := s.Client.Do(req)
	if err != nil {
		return fmt.Errorf("GitHubUploadAssets.Dry: %s", err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("GitHubUploadAssets.Dry: %s", newHTTPError(res))
	}

	return nil
}

// Run executes the step
func (s *GitHubUploadAssets) Run(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Run(ctx)
	}

	s.Result.Assets = make([]GitHubAsset, 0)

	doneCh := make(chan error, len(s.AssetFiles))

	for _, file := range s.AssetFiles {
		go func(file string) {
			assetPath := filepath.Clean(file)
			assetName := filepath.Base(assetPath)

			method := "POST"
			re := regexp.MustCompile(`\{\?[0-9A-Za-z_,]+\}`)
			url := re.ReplaceAllLiteralString(s.ReleaseUploadURL, "")
			url = fmt.Sprintf("%s?name=%s", url, netURL.QueryEscape(assetName))

			content, err := getUploadContent(assetPath)
			if err != nil {
				doneCh <- err
				return
			}
			defer content.Body.Close()

			req, err := createGitHubRequest(ctx, s.Token, method, url, content.Body)
			if err != nil {
				doneCh <- err
				return
			}

			req.Header.Set("Content-Type", content.MIMEType)
			req.ContentLength = content.Length

			res, err := s.Client.Do(req)
			if err != nil {
				doneCh <- err
				return
			}
			defer res.Body.Close()

			if res.StatusCode != 201 {
				doneCh <- newHTTPError(res)
				return
			}

			asset := GitHubAsset{}
			err = json.NewDecoder(res.Body).Decode(&asset)
			if err != nil {
				doneCh <- err
				return
			}

			s.Result.Assets = append(s.Result.Assets, asset)
			doneCh <- nil
		}(file)
	}

	for range s.AssetFiles {
		if err := <-doneCh; err != nil {
			return fmt.Errorf("GitHubUploadAssets.Run: %s", err)
		}
	}

	return nil
}

// Revert reverts back an executed step
func (s *GitHubUploadAssets) Revert(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Revert(ctx)
	}

	for _, asset := range s.Result.Assets {
		method := "DELETE"
		url := fmt.Sprintf("%s/repos/%s/releases/assets/%d", s.BaseURL, s.Repo, asset.ID)
		req, err := createGitHubRequest(ctx, s.Token, method, url, nil)
		if err != nil {
			return err
		}

		res, err := s.Client.Do(req)
		if err != nil {
			return fmt.Errorf("GitHubUploadAssets.Revert: %s", err)
		}
		defer res.Body.Close()

		if res.StatusCode != 204 {
			return fmt.Errorf("GitHubUploadAssets.Revert: %s", newHTTPError(res))
		}
	}

	return nil
}

type uploadContent struct {
	Body     io.ReadCloser
	Length   int64
	MIMEType string
}

func getUploadContent(filepath string) (*uploadContent, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	// Read the first 512 bytes of file to determine the mime type of asset
	buff := make([]byte, 512)
	_, err = file.Read(buff)
	if err != nil {
		return nil, err
	}

	// Determine mime type of asset
	// http.DetectContentType will return "application/octet-stream" if it cannot determine a more specific one
	mimeType := http.DetectContentType(buff)

	// Reset the offset back to the beginning of the file
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}

	return &uploadContent{
		Body:     file,
		Length:   stat.Size(),
		MIMEType: mimeType,
	}, nil
}

// GitHubDownloadAsset downloads an asset file and writes to a local file.
type GitHubDownloadAsset struct {
	Mock      Step
	Client    *http.Client
	Token     string
	BaseURL   string
	Repo      string
	Tag       string
	AssetName string
	Filepath  string
	Result    struct {
		Size int64
	}
}

func (s *GitHubDownloadAsset) makeRequest(ctx context.Context) (io.ReadCloser, error) {
	url := fmt.Sprintf("%s/%s/releases/download/%s/%s", s.BaseURL, s.Repo, s.Tag, s.AssetName)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)
	req.Header.Set("Authorization", fmt.Sprintf("token %s", s.Token))
	req.Header.Set("User-Agent", "cherry") // ref: https://developer.github.com/v3/#user-agent-required

	res, err := s.Client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		return nil, newHTTPError(res)
	}

	return res.Body, nil
}

// Dry is a dry run of the step.
func (s *GitHubDownloadAsset) Dry(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Dry(ctx)
	}

	body, err := s.makeRequest(ctx)
	if err != nil {
		return fmt.Errorf("GitHubDownloadAsset.Dry: %s", err)
	}
	defer body.Close()

	return nil
}

// Run executes the step.
func (s *GitHubDownloadAsset) Run(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Run(ctx)
	}

	body, err := s.makeRequest(ctx)
	if err != nil {
		return fmt.Errorf("GitHubDownloadAsset.Run: %s", err)
	}
	defer body.Close()

	file, err := os.OpenFile(s.Filepath, os.O_WRONLY, 0755)
	if err != nil {
		return fmt.Errorf("GitHubDownloadAsset.Run: %s", err)
	}

	size, err := io.Copy(file, body)
	if err != nil {
		return fmt.Errorf("GitHubDownloadAsset.Run: %s", err)
	}

	s.Result.Size = size

	return nil
}

// Revert reverts back an executed step.
func (s *GitHubDownloadAsset) Revert(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Revert(ctx)
	}

	err := os.Remove(s.Filepath)
	if err != nil {
		return fmt.Errorf("GitHubDownloadAsset.Revert: %s", err)
	}

	return nil
}
