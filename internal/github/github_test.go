package github

import (
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
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
	contextWithTimeout, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

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
		body               string
		expectedError      string
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name:          "InvalidRequest",
			token:         "github-token",
			repo:          "username/repo",
			ctx:           context.Background(),
			method:        "GET",
			endpoint:      " ",
			contentType:   "",
			body:          "",
			expectedError: `invalid character " " in host name`,
		},
		{
			name:          "DoError",
			token:         "github-token",
			repo:          "username/repo",
			ctx:           context.Background(),
			method:        "GET",
			endpoint:      "no_slash",
			contentType:   "",
			body:          "",
			expectedError: "invalid URL port",
		},
		{
			name:          "ContextTimeout",
			token:         "github-token",
			repo:          "username/repo",
			ctx:           contextWithTimeout,
			method:        "GET",
			endpoint:      "/users/moorara",
			contentType:   "",
			body:          "",
			expectedError: "context deadline exceeded",
		},
		{
			name:               "Success200",
			mockStatusCode:     200,
			mockBody:           `{ "login": "moorara" }`,
			token:              "github-token",
			repo:               "username/repo",
			ctx:                context.Background(),
			method:             "GET",
			endpoint:           "/users/moorara",
			contentType:        "",
			body:               "",
			expectedStatusCode: 200,
			expectedBody:       `{ "login": "moorara" }`,
		},
		{
			name:               "Success201",
			mockStatusCode:     201,
			mockBody:           `{ "id": 1, "name": "1.0.0", "tag_name": "v1.0.0" }`,
			token:              "github-token",
			repo:               "username/repo",
			ctx:                context.Background(),
			method:             "POST",
			endpoint:           "/repos/moorara/cherry/releases",
			contentType:        "application/json",
			body:               `{ "name": "1.0.0", "tag_name": "v1.0.0" }`,
			expectedStatusCode: 201,
			expectedBody:       `{ "id": 1, "name": "1.0.0", "tag_name": "v1.0.0" }`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "token "+tc.token, r.Header.Get("Authorization"))
				assert.Equal(t, "moorara/cherry", r.Header.Get("User-Agent"))
				assert.Equal(t, "application/vnd.github.v3+json", r.Header.Get("Accept"))
				assert.Equal(t, "deflate, gzip;q=1.0, *;q=0.5", r.Header.Get("Accept-Encoding"))
				if tc.contentType != "" && tc.body != "" {
					assert.Equal(t, tc.contentType, r.Header.Get("Content-Type"))
				}

				time.Sleep(2 * time.Millisecond)
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

			var body io.Reader
			if tc.body != "" {
				body = strings.NewReader(tc.body)
			}

			url := ts.URL + tc.endpoint
			res, err := gh.makeRequest(tc.ctx, tc.method, url, tc.contentType, body)

			if tc.expectedError != "" {
				assert.Contains(t, err.Error(), tc.expectedError)
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
	contextWithTimeout, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

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
			name:           "ContextTimeout",
			mockStatusCode: 200,
			mockBody:       `{}`,
			token:          "github-token",
			repo:           "username/repo",
			ctx:            contextWithTimeout,
			branch:         "master",
			enabled:        true,
			expectedError:  "context deadline exceeded",
		},
		{
			name:           "BadStatusCode",
			mockStatusCode: 400,
			mockBody:       `{ "enabled": true }`,
			token:          "github-token",
			repo:           "username/repo",
			ctx:            context.Background(),
			branch:         "master",
			enabled:        true,
			expectedError:  "unexpected status code 400",
		},
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
				time.Sleep(2 * time.Millisecond)
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
				apiAddr: ts.URL,
				token:   tc.token,
				repo:    tc.repo,
			}

			err := gh.BranchProtectionForAdmin(tc.ctx, tc.branch, tc.enabled)

			if tc.expectedError != "" {
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCreateRelease(t *testing.T) {
	contextWithTimeout, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

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
			name:           "ContextTimeout",
			mockStatusCode: 200,
			mockBody:       `{}`,
			token:          "github-token",
			repo:           "username/repo",
			ctx:            contextWithTimeout,
			branch:         "master",
			version:        "0.1.0",
			changelog:      "change log description",
			draf:           false,
			prerelease:     false,
			expectedError:  "context deadline exceeded",
		},
		{
			name:           "BadStatusCode",
			mockStatusCode: 400,
			mockBody:       "",
			token:          "github-token",
			repo:           "username/repo",
			ctx:            context.Background(),
			branch:         "master",
			version:        "0.1.0",
			changelog:      "change log description",
			draf:           false,
			prerelease:     false,
			expectedError:  "unexpected status code 400",
		},
		{
			name:           "InvalidResponse",
			mockStatusCode: 201,
			mockBody:       `{ invalid }`,
			token:          "github-token",
			repo:           "username/repo",
			ctx:            context.Background(),
			branch:         "master",
			version:        "0.1.0",
			changelog:      "change log description",
			draf:           false,
			prerelease:     false,
			expectedError:  "invalid character 'i' looking for beginning of object key string",
		},
		{
			name:           "Success",
			mockStatusCode: 201,
			mockBody:       `{ "id": 1, "name": "0.1.0", "tag_name": "v0.1.0" }`,
			token:          "github-token",
			repo:           "username/repo",
			ctx:            context.Background(),
			branch:         "master",
			version:        "0.1.0",
			changelog:      "change log description",
			draf:           false,
			prerelease:     false,
			expectedRelease: &Release{
				ID:         1,
				Name:       "0.1.0",
				Draft:      false,
				Prerelease: false,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(2 * time.Millisecond)
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
				apiAddr: ts.URL,
				token:   tc.token,
				repo:    tc.repo,
			}

			release, err := gh.CreateRelease(tc.ctx, tc.branch, tc.version, tc.changelog, tc.draf, tc.prerelease)

			if tc.expectedError != "" {
				assert.Contains(t, err.Error(), tc.expectedError)
				assert.Nil(t, release)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedRelease, release)
			}
		})
	}
}

func TestUploadAssets(t *testing.T) {
	contextWithTimeout, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	tests := []struct {
		name                      string
		mockGetReleaseStatusCode  int
		mockGetReleaseBody        string
		mockUploadAssetStatusCode int
		mockUploadAssetBody       string
		token                     string
		repo                      string
		ctx                       context.Context
		version                   string
		assets                    []string
		expectedError             string
	}{
		{
			name:                     "GetReleaseContextTimeout",
			mockGetReleaseStatusCode: 200,
			mockGetReleaseBody:       `{}`,
			token:                    "github-token",
			repo:                     "username/repo",
			ctx:                      contextWithTimeout,
			version:                  "0.1.0",
			assets:                   []string{},
			expectedError:            "context deadline exceeded",
		},
		{
			name:                     "GetReleaseBadStatusCode",
			mockGetReleaseStatusCode: 400,
			mockGetReleaseBody:       "",
			token:                    "github-token",
			repo:                     "username/repo",
			ctx:                      context.Background(),
			version:                  "0.1.0",
			assets:                   []string{},
			expectedError:            "unexpected status code 400",
		},
		{
			name:                     "GetReleaseInvalidResponse",
			mockGetReleaseStatusCode: 200,
			mockGetReleaseBody:       `{ invalid }`,
			token:                    "github-token",
			repo:                     "username/repo",
			ctx:                      context.Background(),
			version:                  "0.1.0",
			assets:                   []string{},
			expectedError:            "invalid character 'i' looking for beginning of object key string",
		},
		{
			name:                     "NoAsset",
			mockGetReleaseStatusCode: 200,
			mockGetReleaseBody:       `{ "id": 1 }`,
			token:                    "github-token",
			repo:                     "username/repo",
			ctx:                      context.Background(),
			version:                  "0.1.0",
			assets:                   []string{"./test/nil"},
			expectedError:            "open test/nil: no such file or directory",
		},
		{
			name:                     "EmptyAsset",
			mockGetReleaseStatusCode: 200,
			mockGetReleaseBody:       `{ "id": 1 }`,
			token:                    "github-token",
			repo:                     "username/repo",
			ctx:                      context.Background(),
			version:                  "0.1.0",
			assets:                   []string{"./test/empty"},
			expectedError:            "EOF",
		},
		{
			name:                      "UploadAssetRequestError",
			mockGetReleaseStatusCode:  200,
			mockGetReleaseBody:        `{ "id": 1 }`,
			mockUploadAssetStatusCode: 0,
			mockUploadAssetBody:       "",
			token:                     "github-token",
			repo:                      "username/repo",
			ctx:                       context.Background(),
			version:                   "0.1.0",
			assets:                    []string{"./test/asset"},
			expectedError:             "read: connection reset by peer",
		},
		{
			name:                      "UploadAssetBadStatusCode",
			mockGetReleaseStatusCode:  200,
			mockGetReleaseBody:        `{ "id": 1 }`,
			mockUploadAssetStatusCode: 500,
			mockUploadAssetBody:       "",
			token:                     "github-token",
			repo:                      "username/repo",
			ctx:                       context.Background(),
			version:                   "0.1.0",
			assets:                    []string{"./test/asset"},
			expectedError:             "unexpected status code 500",
		},
		{
			name:                      "Successful",
			mockGetReleaseStatusCode:  200,
			mockGetReleaseBody:        `{ "id": 1 }`,
			mockUploadAssetStatusCode: 201,
			mockUploadAssetBody:       `{}`,
			token:                     "github-token",
			repo:                      "username/repo",
			ctx:                       context.Background(),
			version:                   "0.1.0",
			assets:                    []string{"./test/asset"},
			expectedError:             "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := mux.NewRouter()
			r.Methods("GET").Path("/repos/{owner}/{repo}/releases/tags/{tag}").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(2 * time.Millisecond)
				w.WriteHeader(tc.mockGetReleaseStatusCode)
				w.Write([]byte(tc.mockGetReleaseBody))
			})

			r.Methods("POST").Path("/repos/{owner}/{repo}/releases/{id}/assets").Queries("name", "{name}").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(2 * time.Millisecond)
				w.WriteHeader(tc.mockUploadAssetStatusCode)
				w.Write([]byte(tc.mockUploadAssetBody))
			})

			ts := httptest.NewServer(r)
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

			if tc.expectedError != "" {
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
