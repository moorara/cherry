package service

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
)

const (
	apiAddr    = "https://api.github.com"
	uploadAddr = "https://uploads.github.com"

	acceptType = "application/vnd.github.v3+json"
	userAgent  = "moorara/cherry"
)

type (
	// Github is the interface for API calls to GitHub
	Github interface {
		BranchProtectionForAdmin(ctx context.Context, repo, branch string, enabled bool) error
		CreateRelease(ctx context.Context, repo, branch string, version SemVer, description string, draf, prerelease bool) (*Release, error)
		GetRelease(ctx context.Context, repo string, version SemVer) (*Release, error)
		UploadAssets(ctx context.Context, repo string, version SemVer, assets []string) error
	}

	github struct {
		client     *http.Client
		apiAddr    string
		uploadAddr string
		authHeader string
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
	}

	uploadContent struct {
		Body     io.ReadCloser
		Length   int64
		MIMEType string
	}
)

// NewGithub creates a new Github instance
func NewGithub(timeout time.Duration, token string) Github {
	transport := &http.Transport{}
	client := &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}

	return &github{
		client:     client,
		apiAddr:    apiAddr,
		uploadAddr: uploadAddr,
		authHeader: "token " + token,
	}
}

func (gh *github) getUploadContent(filepath string) (*uploadContent, error) {
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
	// SEEK_SET: seek relative to the origin of the file
	_, err = file.Seek(0, os.SEEK_SET)
	if err != nil {
		return nil, err
	}

	return &uploadContent{
		Body:     file,
		Length:   stat.Size(),
		MIMEType: mimeType,
	}, nil
}

func (gh *github) BranchProtectionForAdmin(ctx context.Context, repo, branch string, enabled bool) error {
	var method string
	var statusCode int

	if enabled {
		method = "POST"
		statusCode = 200
	} else {
		method = "DELETE"
		statusCode = 204
	}

	url := fmt.Sprintf("%s/repos/%s/branches/%s/protection/enforce_admins", gh.apiAddr, repo, branch)

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", gh.authHeader)
	req.Header.Set("Accept", acceptType)
	req.Header.Set("User-Agent", userAgent) // ref: https://developer.github.com/v3/#user-agent-required

	req = req.WithContext(ctx)

	res, err := gh.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != statusCode {
		return util.NewHTTPError(res)
	}

	return nil
}

func (gh *github) CreateRelease(ctx context.Context, repo, branch string, version SemVer, description string, draf, prerelease bool) (*Release, error) {
	method := "POST"
	url := fmt.Sprintf("%s/repos/%s/releases", gh.apiAddr, repo)
	reqBody := releaseReq{
		Name:       version.Version(),
		TagName:    version.GitTag(),
		Target:     branch,
		Draft:      draf,
		Prerelease: prerelease,
		Body:       description,
	}

	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(reqBody)

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", gh.authHeader)
	req.Header.Set("Accept", acceptType)
	req.Header.Set("User-Agent", userAgent) // ref: https://developer.github.com/v3/#user-agent-required
	req.Header.Set("Content-Type", "application/json")

	req = req.WithContext(ctx)

	res, err := gh.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 201 {
		return nil, util.NewHTTPError(res)
	}

	release := new(Release)
	err = json.NewDecoder(res.Body).Decode(release)
	if err != nil {
		return nil, err
	}

	return release, nil
}

func (gh *github) GetRelease(ctx context.Context, repo string, version SemVer) (*Release, error) {
	method := "GET"
	url := fmt.Sprintf("%s/repos/%s/releases/tags/%s", gh.apiAddr, repo, version.GitTag())

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", gh.authHeader)
	req.Header.Set("Accept", acceptType)
	req.Header.Set("User-Agent", userAgent) // ref: https://developer.github.com/v3/#user-agent-required

	req = req.WithContext(ctx)

	res, err := gh.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, util.NewHTTPError(res)
	}

	release := new(Release)
	err = json.NewDecoder(res.Body).Decode(release)
	if err != nil {
		return nil, err
	}

	return release, nil
}

func (gh *github) UploadAssets(ctx context.Context, repo string, version SemVer, assets []string) error {
	release, err := gh.GetRelease(ctx, repo, version)
	if err != nil {
		return err
	}

	for _, asset := range assets {
		assetPath := filepath.Clean(asset)
		assetName := filepath.Base(assetPath)

		method := "POST"
		url := fmt.Sprintf("%s/repos/%s/releases/%d/assets?name=%s", gh.uploadAddr, repo, release.ID, netURL.QueryEscape(assetName))

		content, err := gh.getUploadContent(assetPath)
		if err != nil {
			return err
		}

		req, err := http.NewRequest(method, url, content.Body)
		if err != nil {
			return err
		}

		req.Header.Set("Authorization", gh.authHeader)
		req.Header.Set("Accept", acceptType)
		req.Header.Set("User-Agent", userAgent) // ref: https://developer.github.com/v3/#user-agent-required
		req.Header.Set("Content-Type", content.MIMEType)
		req.ContentLength = content.Length

		req = req.WithContext(ctx)

		res, err := gh.client.Do(req)
		if err != nil {
			return err
		}

		if res.StatusCode != 201 {
			return util.NewHTTPError(res)
		}

		content.Body.Close()
		res.Body.Close()
	}

	return nil
}
