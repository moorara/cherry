package github

import (
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/moorara/cherry/pkg/log"
	"github.com/moorara/cherry/pkg/metrics"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		timeout time.Duration
		token   string
		repo    string
	}{
		{
			"OK",
			2 * time.Second,
			"github_token",
			"username/repo",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			logger := log.NewNopLogger()
			metrics := metrics.New("test")
			gh := New(logger, metrics, tc.timeout, tc.token, tc.repo)

			assert.NotNil(t, gh)
		})
	}
}

func TestMakeRequest(t *testing.T) {
	tests := []struct {
		name               string
		mockStatusCode     int
		mockBody           string
		token              string
		repo               string
		ctx                context.Context
		method             string
		endpoint           string
		contentType        string
		body               io.Reader
		expectedError      string
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name:               "Success200",
			mockStatusCode:     200,
			mockBody:           `{}`,
			token:              "github-token",
			repo:               "username/repo",
			ctx:                context.Background(),
			method:             "GET",
			endpoint:           "/users/moorara",
			contentType:        "application/json",
			body:               nil,
			expectedError:      "",
			expectedStatusCode: 200,
			expectedBody:       `{}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.mockStatusCode)
				w.Write([]byte(tc.mockBody))
			}))
			defer ts.Close()

			logger := log.NewNopLogger()
			metrics := metrics.New("test")
			client := &http.Client{}
			gh := &github{
				logger:  logger,
				metrics: metrics,
				client:  client,
				token:   tc.token,
				repo:    tc.repo,
			}

			url := ts.URL + tc.endpoint
			res, err := gh.makeRequest(tc.ctx, tc.method, url, tc.contentType, tc.body)

			if tc.expectedError != "" {
				assert.Equal(t, tc.expectedError, err.Error())
				assert.Nil(t, res)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, res)

				data, err := ioutil.ReadAll(res.Body)
				assert.NoError(t, err)
				res.Body.Close()

				assert.Equal(t, tc.expectedStatusCode, res.StatusCode)
				assert.Equal(t, tc.expectedBody, string(data))
			}
		})
	}
}

func TestBranchProtectionForAdmin(t *testing.T) {
	tests := []struct {
		name           string
		mockStatusCode int
		mockBody       string
		token          string
		repo           string
		ctx            context.Context
		branch         string
		enabled        bool
		expectedError  string
	}{
		{
			name:           "Enable",
			mockStatusCode: 200,
			mockBody:       `{ "enabled": true }`,
			token:          "github-token",
			repo:           "username/repo",
			ctx:            context.Background(),
			branch:         "master",
			enabled:        true,
			expectedError:  "",
		},
		{
			name:           "Disable",
			mockStatusCode: 204,
			mockBody:       `{ "enabled": false }`,
			token:          "github-token",
			repo:           "username/repo",
			ctx:            context.Background(),
			branch:         "master",
			enabled:        false,
			expectedError:  "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.mockStatusCode)
				w.Write([]byte(tc.mockBody))
			}))
			defer ts.Close()

			logger := log.NewNopLogger()
			metrics := metrics.New("test")
			client := &http.Client{}
			gh := &github{
				logger:     logger,
				metrics:    metrics,
				client:     client,
				apiAddr:    ts.URL,
				uploadAddr: ts.URL,
				token:      tc.token,
				repo:       tc.repo,
			}

			err := gh.BranchProtectionForAdmin(tc.ctx, tc.branch, tc.enabled)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}

func TestCreateRelease(t *testing.T) {
	tests := []struct {
		name            string
		mockStatusCode  int
		mockBody        string
		token           string
		repo            string
		ctx             context.Context
		branch          string
		version         string
		changelog       string
		draf            bool
		prerelease      bool
		expectedError   string
		expectedRelease *Release
	}{
		{
			name:           "",
			mockStatusCode: 201,
			mockBody:       `{ "id": "aaaa", "name": "0.1.0", "draft": false, "prerelease": false }`,
			token:          "github-token",
			repo:           "username/repo",
			ctx:            context.Background(),
			branch:         "master",
			version:        "0.1.0",
			changelog:      "change log",
			draf:           false,
			prerelease:     false,
			expectedError:  "",
			expectedRelease: &Release{
				ID:         "aaaa",
				Name:       "0.1.0",
				Draft:      false,
				Prerelease: false,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.mockStatusCode)
				w.Write([]byte(tc.mockBody))
			}))
			defer ts.Close()

			logger := log.NewNopLogger()
			metrics := metrics.New("test")
			client := &http.Client{}
			gh := &github{
				logger:     logger,
				metrics:    metrics,
				client:     client,
				apiAddr:    ts.URL,
				uploadAddr: ts.URL,
				token:      tc.token,
				repo:       tc.repo,
			}

			_, err := gh.CreateRelease(tc.ctx, tc.branch, tc.version, tc.changelog, tc.draf, tc.prerelease)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}

func TestUploadAssets(t *testing.T) {
	tests := []struct {
		name           string
		mockStatusCode int
		mockBody       string
		token          string
		repo           string
		ctx            context.Context
		version        string
		assets         []string
		expectedError  string
	}{}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.mockStatusCode)
				w.Write([]byte(tc.mockBody))
			}))
			defer ts.Close()

			logger := log.NewNopLogger()
			metrics := metrics.New("test")
			client := &http.Client{}
			gh := &github{
				logger:     logger,
				metrics:    metrics,
				client:     client,
				apiAddr:    ts.URL,
				uploadAddr: ts.URL,
				token:      tc.token,
				repo:       tc.repo,
			}

			err := gh.UploadAssets(tc.ctx, tc.version, tc.assets)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}
