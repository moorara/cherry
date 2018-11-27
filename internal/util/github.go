package util

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/moorara/cherry/pkg/log"
	"github.com/moorara/cherry/pkg/metrics"
)

type (
	// Github is the interface for API calls to GitHub
	Github interface {
		BranchProtectionForAdmin(ctx context.Context, branch string, enabled bool) error
		CreateRelease(ctx context.Context, branch, version, changelog string, draf, prerelease bool) error
	}

	github struct {
		logger  *log.Logger
		metrics *metrics.Metrics
		client  *http.Client
		token   string
		repo    string
	}

	githubRelease struct {
		Name       string `json:"name"`
		TagName    string `json:"tag_name"`
		Target     string `json:"target_commitish"`
		Draft      bool   `json:"draft"`
		Prerelease bool   `json:"prerelease"`
		Body       string `json:"body"`
	}
)

// NewGithub creates a new Github instance
func NewGithub(logger *log.Logger, metrics *metrics.Metrics, timeout time.Duration, token, repo string) Github {
	transport := &http.Transport{}
	client := &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}

	return &github{
		logger:  logger,
		metrics: metrics,
		token:   token,
		repo:    repo,
		client:  client,
	}
}

func (g *github) makeRequest(ctx context.Context, method, url string, body interface{}) (*http.Response, error) {
	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(body)
	if err != nil {
		g.logger.Error("message", "Error on encoding request body.", "error", err, "method", method, "url", url, "body", body)
		return nil, err
	}

	req, err := http.NewRequest(method, url, buf)
	if err != nil {
		g.logger.Error("message", "Error on creating request.", "error", err, "method", method, "url", url, "body", body)
		return nil, err
	}

	req.Header.Set("Authorization", "token "+g.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	res, err := g.client.Do(req)
	if err != nil {
		g.logger.Error("message", "Error on making request.", "error", err, "method", method, "url", url, "body", body)
		return nil, err
	}

	return res, nil
}

// BranchProtectionForAdmin enables or disables a branch protection for admin users/tokens
func (g *github) BranchProtectionForAdmin(ctx context.Context, branch string, enabled bool) error {
	var method string
	var statusCode int

	if enabled {
		method = "POST"
		statusCode = 200
	} else {
		method = "DELETE"
		statusCode = 204
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/branches/%s/protection/enforce_admins", g.repo, branch)
	res, err := g.makeRequest(ctx, method, url, nil)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != statusCode {
		err := fmt.Errorf("unexpected status code %d", res.StatusCode)
		g.logger.Error("message", "Error on enabling branch protection for admins.", "error", err, "method", method, "url", url)
		return err
	}

	return nil
}

func (g *github) CreateRelease(ctx context.Context, branch, version, changelog string, draf, prerelease bool) error {
	method := "POST"
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases", g.repo)
	reqBody := githubRelease{
		Name:       version,
		TagName:    "v" + version,
		Target:     branch,
		Draft:      draf,
		Prerelease: prerelease,
		Body:       fmt.Sprintf("$comment\n\n%s", changelog),
	}

	res, err := g.makeRequest(ctx, method, url, reqBody)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 201 {
		err := fmt.Errorf("unexpected status code %d", res.StatusCode)
		g.logger.Error("message", "Error on creating release.", "error", err, "method", method, "url", url, "reqBody", reqBody)
		return err
	}

	return nil
}
