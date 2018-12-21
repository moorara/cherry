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
	"github.com/moorara/cherry/internal/service/semver"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		timeout time.Duration
		token   string
	}{
		{
			"OK",
			2 * time.Second,
			"github_token",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			github := New(tc.timeout, tc.token)
			assert.NotNil(t, github)
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

			client := &http.Client{}
			gh := &github{
				client: client,
				token:  tc.token,
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
	tests := []struct {
		name           string
		mockAPI        bool
		mockStatusCode int
		mockBody       string
		token          string
		ctx            context.Context
		repo           string
		branch         string
		enabled        bool
		expectedError  string
	}{
		{
			name:          "RequestError",
			mockAPI:       false,
			token:         "github-token",
			ctx:           context.Background(),
			repo:          "username/repo",
			branch:        "master",
			enabled:       true,
			expectedError: "unsupported protocol scheme",
		},
		{
			name:           "BadStatusCode",
			mockAPI:        true,
			mockStatusCode: 400,
			mockBody:       `{ "enabled": true }`,
			token:          "github-token",
			ctx:            context.Background(),
			repo:           "username/repo",
			branch:         "master",
			enabled:        true,
			expectedError:  "POST /repos/username/repo/branches/master/protection/enforce_admins 400",
		},
		{
			name:           "Enable",
			mockAPI:        true,
			mockStatusCode: 200,
			mockBody:       `{ "enabled": true }`,
			token:          "github-token",
			ctx:            context.Background(),
			repo:           "username/repo",
			branch:         "master",
			enabled:        true,
			expectedError:  "",
		},
		{
			name:           "Disable",
			mockAPI:        true,
			mockStatusCode: 204,
			mockBody:       `{ "enabled": false }`,
			token:          "github-token",
			ctx:            context.Background(),
			repo:           "username/repo",
			branch:         "master",
			enabled:        false,
			expectedError:  "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := &http.Client{}
			gh := &github{
				client: client,
				token:  tc.token,
			}

			if tc.mockAPI {
				r := mux.NewRouter()
				r.Methods("POST", "DELETE").Path("/repos/{owner}/{repo}/branches/{branch}/protection/enforce_admins").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(tc.mockStatusCode)
					w.Write([]byte(tc.mockBody))
				})

				ts := httptest.NewServer(r)
				defer ts.Close()

				gh.apiAddr = ts.URL
			}

			err := gh.BranchProtectionForAdmin(tc.ctx, tc.repo, tc.branch, tc.enabled)

			if tc.expectedError != "" {
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCreateRelease(t *testing.T) {
	tests := []struct {
		name            string
		mockAPI         bool
		mockStatusCode  int
		mockBody        string
		token           string
		ctx             context.Context
		repo            string
		branch          string
		version         semver.SemVer
		description     string
		draf            bool
		prerelease      bool
		expectedError   string
		expectedRelease *Release
	}{
		{
			name:          "RequestError",
			mockAPI:       false,
			token:         "github-token",
			ctx:           context.Background(),
			repo:          "username/repo",
			branch:        "master",
			version:       semver.SemVer{Major: 0, Minor: 1, Patch: 0},
			description:   "release description",
			draf:          false,
			prerelease:    false,
			expectedError: "unsupported protocol scheme",
		},
		{
			name:           "BadStatusCode",
			mockAPI:        true,
			mockStatusCode: 400,
			mockBody:       "",
			token:          "github-token",
			ctx:            context.Background(),
			repo:           "username/repo",
			branch:         "master",
			version:        semver.SemVer{Major: 0, Minor: 1, Patch: 0},
			description:    "release description",
			draf:           false,
			prerelease:     false,
			expectedError:  "POST /repos/username/repo/releases 400",
		},
		{
			name:           "InvalidResponse",
			mockAPI:        true,
			mockStatusCode: 201,
			mockBody:       `{ invalid }`,
			token:          "github-token",
			ctx:            context.Background(),
			repo:           "username/repo",
			branch:         "master",
			version:        semver.SemVer{Major: 0, Minor: 1, Patch: 0},
			description:    "release description",
			draf:           false,
			prerelease:     false,
			expectedError:  "invalid character 'i' looking for beginning of object key string",
		},
		{
			name:           "Success",
			mockAPI:        true,
			mockStatusCode: 201,
			mockBody:       `{ "id": 1, "name": "0.1.0", "tag_name": "v0.1.0" }`,
			token:          "github-token",
			ctx:            context.Background(),
			repo:           "username/repo",
			branch:         "master",
			version:        semver.SemVer{Major: 0, Minor: 1, Patch: 0},
			description:    "release description",
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
			client := &http.Client{}
			gh := &github{
				client: client,
				token:  tc.token,
			}

			if tc.mockAPI {
				r := mux.NewRouter()
				r.Methods("POST").Path("/repos/{owner}/{repo}/releases").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(tc.mockStatusCode)
					w.Write([]byte(tc.mockBody))
				})

				ts := httptest.NewServer(r)
				defer ts.Close()

				gh.apiAddr = ts.URL
			}

			release, err := gh.CreateRelease(tc.ctx, tc.repo, tc.branch, tc.version, tc.description, tc.draf, tc.prerelease)

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
	tests := []struct {
		name                      string
		mockAPI                   bool
		mockGetReleaseStatusCode  int
		mockGetReleaseBody        string
		mockUploadAPI             bool
		mockUploadAssetStatusCode int
		mockUploadAssetBody       string
		token                     string
		ctx                       context.Context
		repo                      string
		version                   semver.SemVer
		assets                    []string
		expectedError             string
	}{
		{
			name:          "GetReleaseRequestError",
			mockAPI:       false,
			token:         "github-token",
			ctx:           context.Background(),
			repo:          "username/repo",
			version:       semver.SemVer{Major: 0, Minor: 1, Patch: 0},
			assets:        []string{},
			expectedError: "unsupported protocol scheme",
		},
		{
			name:                     "GetReleaseBadStatusCode",
			mockAPI:                  true,
			mockGetReleaseStatusCode: 400,
			mockGetReleaseBody:       "",
			token:                    "github-token",
			ctx:                      context.Background(),
			repo:                     "username/repo",
			version:                  semver.SemVer{Major: 0, Minor: 1, Patch: 0},
			assets:                   []string{},
			expectedError:            "GET /repos/username/repo/releases/tags/v0.1.0 400",
		},
		{
			name:                     "GetReleaseInvalidResponse",
			mockAPI:                  true,
			mockGetReleaseStatusCode: 200,
			mockGetReleaseBody:       `{ invalid }`,
			token:                    "github-token",
			ctx:                      context.Background(),
			repo:                     "username/repo",
			version:                  semver.SemVer{Major: 0, Minor: 1, Patch: 0},
			assets:                   []string{},
			expectedError:            "invalid character 'i' looking for beginning of object key string",
		},
		{
			name:                     "NoAsset",
			mockAPI:                  true,
			mockGetReleaseStatusCode: 200,
			mockGetReleaseBody:       `{ "id": 1 }`,
			token:                    "github-token",
			ctx:                      context.Background(),
			repo:                     "username/repo",
			version:                  semver.SemVer{Major: 0, Minor: 1, Patch: 0},
			assets:                   []string{"./test/nil"},
			expectedError:            "open test/nil: no such file or directory",
		},
		{
			name:                     "EmptyAsset",
			mockAPI:                  true,
			mockGetReleaseStatusCode: 200,
			mockGetReleaseBody:       `{ "id": 1 }`,
			token:                    "github-token",
			ctx:                      context.Background(),
			repo:                     "username/repo",
			version:                  semver.SemVer{Major: 0, Minor: 1, Patch: 0},
			assets:                   []string{"./test/empty"},
			expectedError:            "EOF",
		},
		{
			name:                     "UploadAssetRequestError",
			mockAPI:                  true,
			mockGetReleaseStatusCode: 200,
			mockGetReleaseBody:       `{ "id": 1 }`,
			mockUploadAPI:            false,
			token:                    "github-token",
			ctx:                      context.Background(),
			repo:                     "username/repo",
			version:                  semver.SemVer{Major: 0, Minor: 1, Patch: 0},
			assets:                   []string{"./test/asset"},
			expectedError:            "unsupported protocol scheme",
		},
		{
			name:                      "UploadAssetBadStatusCode",
			mockAPI:                   true,
			mockGetReleaseStatusCode:  200,
			mockGetReleaseBody:        `{ "id": 1 }`,
			mockUploadAPI:             true,
			mockUploadAssetStatusCode: 500,
			mockUploadAssetBody:       "",
			token:                     "github-token",
			ctx:                       context.Background(),
			repo:                      "username/repo",
			version:                   semver.SemVer{Major: 0, Minor: 1, Patch: 0},
			assets:                    []string{"./test/asset"},
			expectedError:             "POST /repos/username/repo/releases/1/assets 500",
		},
		{
			name:                      "Successful",
			mockAPI:                   true,
			mockGetReleaseStatusCode:  200,
			mockGetReleaseBody:        `{ "id": 1 }`,
			mockUploadAPI:             true,
			mockUploadAssetStatusCode: 201,
			mockUploadAssetBody:       `{}`,
			token:                     "github-token",
			ctx:                       context.Background(),
			repo:                      "username/repo",
			version:                   semver.SemVer{Major: 0, Minor: 1, Patch: 0},
			assets:                    []string{"./test/asset"},
			expectedError:             "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := &http.Client{}
			gh := &github{
				client: client,
				token:  tc.token,
			}

			if tc.mockAPI {
				r := mux.NewRouter()
				r.Methods("GET").Path("/repos/{owner}/{repo}/releases/tags/{tag}").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(tc.mockGetReleaseStatusCode)
					w.Write([]byte(tc.mockGetReleaseBody))
				})

				ts := httptest.NewServer(r)
				defer ts.Close()

				gh.apiAddr = ts.URL
			}

			if tc.mockUploadAPI {
				r := mux.NewRouter()
				r.Methods("POST").Path("/repos/{owner}/{repo}/releases/{id}/assets").Queries("name", "{name}").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(tc.mockUploadAssetStatusCode)
					w.Write([]byte(tc.mockUploadAssetBody))
				})

				ts := httptest.NewServer(r)
				defer ts.Close()

				gh.uploadAddr = ts.URL
			}

			err := gh.UploadAssets(tc.ctx, tc.repo, tc.version, tc.assets)

			if tc.expectedError != "" {
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
