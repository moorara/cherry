package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestNewGithub(t *testing.T) {
	tests := []struct {
		token string
	}{
		{
			"github_token",
		},
	}

	for _, tc := range tests {
		github := NewGithub(tc.token)
		assert.NotNil(t, github)
	}
}

func TestGithubGetUploadContent(t *testing.T) {
	tests := []struct {
		name string
	}{}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.NotNil(t, tc)
		})
	}
}

func TestGithubBranchProtectionForAdmin(t *testing.T) {
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
			gh := &github{
				client:     &http.Client{},
				authHeader: "token " + tc.token,
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

func TestGithubCreateRelease(t *testing.T) {
	tests := []struct {
		name            string
		mockAPI         bool
		mockStatusCode  int
		mockBody        string
		token           string
		ctx             context.Context
		repo            string
		branch          string
		version         SemVer
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
			version:       SemVer{Major: 0, Minor: 1, Patch: 0},
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
			version:        SemVer{Major: 0, Minor: 1, Patch: 0},
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
			version:        SemVer{Major: 0, Minor: 1, Patch: 0},
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
			version:        SemVer{Major: 0, Minor: 1, Patch: 0},
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
			gh := &github{
				client:     &http.Client{},
				authHeader: "token " + tc.token,
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

func TestGithubGetRelease(t *testing.T) {
	tests := []struct {
		name            string
		mockAPI         bool
		mockStatusCode  int
		mockBody        string
		token           string
		ctx             context.Context
		repo            string
		version         SemVer
		expectedError   string
		expectedRelease *Release
	}{
		{
			name:          "RequestError",
			mockAPI:       false,
			token:         "github-token",
			ctx:           context.Background(),
			repo:          "username/repo",
			version:       SemVer{Major: 0, Minor: 1, Patch: 0},
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
			version:        SemVer{Major: 0, Minor: 1, Patch: 0},
			expectedError:  "GET /repos/username/repo/releases/tags/v0.1.0 400",
		},
		{
			name:           "InvalidResponse",
			mockAPI:        true,
			mockStatusCode: 200,
			mockBody:       `{ invalid }`,
			token:          "github-token",
			ctx:            context.Background(),
			repo:           "username/repo",
			version:        SemVer{Major: 0, Minor: 1, Patch: 0},
			expectedError:  "invalid character 'i' looking for beginning of object key string",
		},
		{
			name:           "Success",
			mockAPI:        true,
			mockStatusCode: 200,
			mockBody:       `{ "id": 1, "name": "v0.1.0", "draft": true, "prerelease": true }`,
			token:          "github-token",
			ctx:            context.Background(),
			repo:           "username/repo",
			version:        SemVer{Major: 0, Minor: 1, Patch: 0},
			expectedRelease: &Release{
				ID:         1,
				Name:       "v0.1.0",
				Draft:      true,
				Prerelease: true,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gh := &github{
				client:     &http.Client{},
				authHeader: "token " + tc.token,
			}

			if tc.mockAPI {
				r := mux.NewRouter()
				r.Methods("GET").Path("/repos/{owner}/{repo}/releases/tags/{tag}").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(tc.mockStatusCode)
					w.Write([]byte(tc.mockBody))
				})

				ts := httptest.NewServer(r)
				defer ts.Close()

				gh.apiAddr = ts.URL
			}

			release, err := gh.GetRelease(tc.ctx, tc.repo, tc.version)

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

func TestGithubUploadAssets(t *testing.T) {
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
		version                   SemVer
		assets                    []string
		expectedError             string
	}{
		{
			name:                     "GetReleaseBadStatusCode",
			mockAPI:                  true,
			mockGetReleaseStatusCode: 400,
			mockGetReleaseBody:       "",
			token:                    "github-token",
			ctx:                      context.Background(),
			repo:                     "username/repo",
			version:                  SemVer{Major: 0, Minor: 1, Patch: 0},
			assets:                   []string{},
			expectedError:            "GET /repos/username/repo/releases/tags/v0.1.0 400",
		},
		{
			name:                     "NoAsset",
			mockAPI:                  true,
			mockGetReleaseStatusCode: 200,
			mockGetReleaseBody:       `{ "id": 1, "name": "v0.1.0" }`,
			token:                    "github-token",
			ctx:                      context.Background(),
			repo:                     "username/repo",
			version:                  SemVer{Major: 0, Minor: 1, Patch: 0},
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
			version:                  SemVer{Major: 0, Minor: 1, Patch: 0},
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
			version:                  SemVer{Major: 0, Minor: 1, Patch: 0},
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
			version:                   SemVer{Major: 0, Minor: 1, Patch: 0},
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
			version:                   SemVer{Major: 0, Minor: 1, Patch: 0},
			assets:                    []string{"./test/asset"},
			expectedError:             "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := &http.Client{}
			gh := &github{
				client:     client,
				authHeader: "token " + tc.token,
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
