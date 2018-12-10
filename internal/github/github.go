package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	netURL "net/url"

	"github.com/moorara/cherry/internal/util"
	"github.com/moorara/cherry/pkg/log"
	"github.com/moorara/cherry/pkg/metrics"
)

const (
	apiAddr         = "https://api.github.com"
	uploadAddr      = "https://uploads.github.com"
	jsonContentType = "application/json"
)

type (
	// Github is the interface for API calls to GitHub
	Github interface {
		BranchProtectionForAdmin(ctx context.Context, branch string, enabled bool) error
		CreateRelease(ctx context.Context, branch, version, changelog string, draf, prerelease bool) (*Release, error)
		UploadAssets(ctx context.Context, version string, assets []string) error
	}

	github struct {
		logger     *log.Logger
		metrics    *metrics.Metrics
		client     *http.Client
		apiAddr    string
		uploadAddr string
		token      string
		repo       string
	}

	releaseReq struct {
		Name       string `json:"name"`
		TagName    string `json:"tag_name"`
		Target     string `json:"target_commitish"`
		Draft      bool   `json:"draft"`
		Prerelease bool   `json:"prerelease"`
		Body       string `json:"body"`
	}

	// Release is the model for GitHub release
	Release struct {
		ID         int    `json:"id"`
		Name       string `json:"name"`
		Draft      bool   `json:"draft"`
		Prerelease bool   `json:"prerelease"`
		URL        string `json:"url"`
		HTMLURL    string `json:"html_url"`
		AssetsURL  string `json:"assets_url"`
		UploadURL  string `json:"upload_url"`
		TarballURL string `json:"tarball_url"`
		ZipballURL string `json:"zipball_url"`
	}
)

// New creates a new Github instance
func New(logger *log.Logger, metrics *metrics.Metrics, timeout time.Duration, token, repo string) Github {
	transport := &http.Transport{}
	client := &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}

	return &github{
		logger:     logger,
		metrics:    metrics,
		client:     client,
		apiAddr:    apiAddr,
		uploadAddr: uploadAddr,
		token:      token,
		repo:       repo,
	}
}

func (gh *github) makeRequest(ctx context.Context, method, url, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		gh.logger.Error("message", "Error on creating request.", "error", err, "method", method, "url", url, "contentType", contentType)
		return nil, err
	}

	req.Header.Set("Authorization", "token "+gh.token)
	req.Header.Set("User-Agent", "moorara/cherry") // ref: https://developer.github.com/v3/#user-agent-required
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Accept-Encoding", "deflate, gzip;q=1.0, *;q=0.5")
	if contentType != "" && body != nil {
		req.Header.Set("Content-Type", contentType)
	}

	res, err := gh.client.Do(req)
	if err != nil {
		gh.logger.Error("message", "Error on making request.", "error", err, "method", method, "url", url, "contentType", contentType)
		return nil, err
	}

	return res, nil
}

func (gh *github) BranchProtectionForAdmin(ctx context.Context, branch string, enabled bool) error {
	var method string
	var statusCode int

	if enabled {
		method = "POST"
		statusCode = 200
	} else {
		method = "DELETE"
		statusCode = 204
	}

	url := fmt.Sprintf("%s/repos/%s/branches/%s/protection/enforce_admins", gh.apiAddr, gh.repo, branch)
	res, err := gh.makeRequest(ctx, method, url, "", nil)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != statusCode {
		err := util.NewHTTPError(res)
		gh.logger.Error("message", "Error on changing branch protection for admins.", "error", err, "method", method, "url", url, "resBody", err.Body)
		return err
	}

	return nil
}

func (gh *github) CreateRelease(ctx context.Context, branch, version, changelog string, draf, prerelease bool) (*Release, error) {
	method := "POST"
	url := fmt.Sprintf("%s/repos/%s/releases", gh.apiAddr, gh.repo)
	reqBody := releaseReq{
		Name:       version,
		TagName:    "v" + version,
		Target:     branch,
		Draft:      draf,
		Prerelease: prerelease,
		Body:       fmt.Sprintf("$comment\n\n%s", changelog),
	}

	buff := new(bytes.Buffer)
	json.NewEncoder(buff).Encode(reqBody)

	res, err := gh.makeRequest(ctx, method, url, jsonContentType, buff)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 201 {
		err := util.NewHTTPError(res)
		gh.logger.Error("message", "Error on creating release.", "error", err, "method", method, "url", url, "reqBody", reqBody, "resBody", err.Body)
		return nil, err
	}

	release := new(Release)
	err = json.NewDecoder(res.Body).Decode(release)
	if err != nil {
		gh.logger.Error("message", "Error on response body.", "error", err, "method", method, "url", url)
		return nil, err
	}

	return release, nil
}

func (gh *github) UploadAssets(ctx context.Context, version string, assets []string) error {
	method := "GET"
	url := fmt.Sprintf("%s/repos/%s/releases/tags/v%s", gh.apiAddr, gh.repo, version)

	res, err := gh.makeRequest(ctx, method, url, "application/json", nil)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		err := util.NewHTTPError(res)
		gh.logger.Error("message", "Error on getting release.", "error", err, "method", method, "url", url, "resBody", err.Body)
		return err
	}

	release := new(Release)
	err = json.NewDecoder(res.Body).Decode(release)
	if err != nil {
		gh.logger.Error("message", "Error on response body.", "error", err, "method", method, "url", url)
		return err
	}

	for _, asset := range assets {
		assetFilePath := filepath.Clean(asset)
		assetFileName := filepath.Base(assetFilePath)

		file, err := os.Open(assetFilePath)
		if err != nil {
			gh.logger.Error("message", "Error on opening asset file.", "error", err)
			return err
		}

		// Read the first 512 bytes of file to determine the mime type of asset
		buff := make([]byte, 512)
		_, err = file.Read(buff)
		if err != nil {
			gh.logger.Error("message", "Error on reading asset file.", "error", err)
			return err
		}

		// Determine mime type of asset
		// http.DetectContentType will return "application/octet-stream" if it cannot determine a more specific one
		contentType := http.DetectContentType(buff)

		// Reset the offset back to the beginning of the file
		// SEEK_SET: seek relative to the origin of the file
		file.Seek(0, os.SEEK_SET)

		method := "POST"
		url := fmt.Sprintf("%s/repos/%s/releases/%d/assets?name=%s", gh.uploadAddr, gh.repo, release.ID, netURL.QueryEscape(assetFileName))
		res, err := gh.makeRequest(ctx, method, url, contentType, file)
		if err != nil {
			return err
		}

		if res.StatusCode != 201 {
			err := util.NewHTTPError(res)
			gh.logger.Error("message", "Error on uploading release asset.", "error", err, "method", method, "url", url, "resBody", err.Body)
			return err
		}

		file.Close()
		res.Body.Close()
	}

	return nil
}
