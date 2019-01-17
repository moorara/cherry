package service

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

const (
	releaseBody1 = `{
		"id": 11111111,
		"name": "0.1.0",
		"tag_name": "v0.1.0",
		"target_commitish": "master",
		"draft": false,
		"prerelease": false,
		"body": "release description",
		"assets": [{
			"id": 1111,
			"name": "app-darwin-amd64",
			"label": "",
			"state": "uploaded",
			"size": 1048576,
			"content_type": "application/octet-stream",
			"browser_download_url": "https://github.com/username/repo/releases/download/v0.1.0/app-darwin-amd64"
		}, {
			"id": 2222,
			"name": "app-linux-amd64",
			"label": "",
			"state": "uploaded",
			"size": 1048576,
			"content_type": "application/octet-stream",
			"browser_download_url": "https://github.com/username/repo/releases/download/v0.1.0/app-linux-amd64"
		}]
	}`

	releaseBody2 = `{
		"id": 22222222,
		"name": "0.2.0",
		"tag_name": "v0.2.0",
		"target_commitish": "master",
		"draft": false,
		"prerelease": false,
		"body": "release description",
		"assets": [{
			"id": 3333,
			"name": "app-darwin-amd64",
			"label": "",
			"state": "uploaded",
			"size": 1048576,
			"content_type": "application/octet-stream",
			"browser_download_url": "https://github.com/username/repo/releases/download/v0.2.0/app-darwin-amd64"
		}, {
			"id": 4444,
			"name": "app-linux-amd64",
			"label": "",
			"state": "uploaded",
			"size": 1048576,
			"content_type": "application/octet-stream",
			"browser_download_url": "https://github.com/username/repo/releases/download/v0.2.0/app-linux-amd64"
		}]
	}`
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

func TestIsTokenValid(t *testing.T) {
	tests := []struct {
		name           string
		authHeader     string
		expectedResult bool
	}{
		{
			name:           "InvalidPrefix",
			authHeader:     "github-token",
			expectedResult: false,
		},
		{
			name:           "InvalidLength",
			authHeader:     "token ",
			expectedResult: false,
		},
		{
			name:           "Success",
			authHeader:     "token github-token",
			expectedResult: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gh := &github{
				authHeader: tc.authHeader,
			}

			result := gh.isTokenValid()

			assert.Equal(t, tc.expectedResult, result)
		})
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
		input           ReleaseInput
		expectedError   string
		expectedRelease *Release
	}{
		{
			name:    "RequestError",
			mockAPI: false,
			token:   "github-token",
			ctx:     context.Background(),
			repo:    "username/repo",
			input: ReleaseInput{
				Name:       "0.1.0",
				TagName:    "v0.1.0",
				Target:     "master",
				Draft:      false,
				Prerelease: false,
				Body:       "release description",
			},
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
			input: ReleaseInput{
				Name:       "0.1.0",
				TagName:    "v0.1.0",
				Target:     "master",
				Draft:      false,
				Prerelease: false,
				Body:       "release description",
			},
			expectedError: "POST /repos/username/repo/releases 400",
		},
		{
			name:           "InvalidResponse",
			mockAPI:        true,
			mockStatusCode: 201,
			mockBody:       `{ invalid }`,
			token:          "github-token",
			ctx:            context.Background(),
			repo:           "username/repo",
			input: ReleaseInput{
				Name:       "0.1.0",
				TagName:    "v0.1.0",
				Target:     "master",
				Draft:      false,
				Prerelease: false,
				Body:       "release description",
			},
			expectedError: "invalid character 'i' looking for beginning of object key string",
		},
		{
			name:           "Success",
			mockAPI:        true,
			mockStatusCode: 201,
			mockBody:       releaseBody1,
			token:          "github-token",
			ctx:            context.Background(),
			repo:           "username/repo",
			input: ReleaseInput{
				Name:       "0.1.0",
				TagName:    "v0.1.0",
				Target:     "master",
				Draft:      false,
				Prerelease: false,
				Body:       "release description",
			},
			expectedRelease: &Release{
				ID:         11111111,
				Name:       "0.1.0",
				TagName:    "v0.1.0",
				Target:     "master",
				Draft:      false,
				Prerelease: false,
				Body:       "release description",
				Assets: []Asset{
					{
						ID:          1111,
						Name:        "app-darwin-amd64",
						Label:       "",
						State:       "uploaded",
						Size:        1048576,
						ContentType: "application/octet-stream",
						DownloadURL: "https://github.com/username/repo/releases/download/v0.1.0/app-darwin-amd64",
					},
					{
						ID:          2222,
						Name:        "app-linux-amd64",
						Label:       "",
						State:       "uploaded",
						Size:        1048576,
						ContentType: "application/octet-stream",
						DownloadURL: "https://github.com/username/repo/releases/download/v0.1.0/app-linux-amd64",
					},
				},
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

			release, err := gh.CreateRelease(tc.ctx, tc.repo, tc.input)

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

func TestGithubEditRelease(t *testing.T) {
	tests := []struct {
		name            string
		mockAPI         bool
		mockStatusCode  int
		mockBody        string
		token           string
		ctx             context.Context
		repo            string
		releaseID       int
		input           ReleaseInput
		expectedError   string
		expectedRelease *Release
	}{
		{
			name:      "RequestError",
			mockAPI:   false,
			token:     "github-token",
			ctx:       context.Background(),
			repo:      "username/repo",
			releaseID: 12345678,
			input: ReleaseInput{
				Name:       "0.1.0",
				TagName:    "v0.1.0",
				Target:     "master",
				Draft:      false,
				Prerelease: false,
				Body:       "release description",
			},
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
			releaseID:      12345678,
			input: ReleaseInput{
				Name:       "0.1.0",
				TagName:    "v0.1.0",
				Target:     "master",
				Draft:      false,
				Prerelease: false,
				Body:       "release description",
			},
			expectedError: "PATCH /repos/username/repo/releases/12345678 400",
		},
		{
			name:           "InvalidResponse",
			mockAPI:        true,
			mockStatusCode: 200,
			mockBody:       `{ invalid }`,
			token:          "github-token",
			ctx:            context.Background(),
			repo:           "username/repo",
			releaseID:      12345678,
			input: ReleaseInput{
				Name:       "0.1.0",
				TagName:    "v0.1.0",
				Target:     "master",
				Draft:      false,
				Prerelease: false,
				Body:       "release description",
			},
			expectedError: "invalid character 'i' looking for beginning of object key string",
		},
		{
			name:           "Success",
			mockAPI:        true,
			mockStatusCode: 200,
			mockBody:       releaseBody1,
			token:          "github-token",
			ctx:            context.Background(),
			repo:           "username/repo",
			releaseID:      12345678,
			input: ReleaseInput{
				Name:       "0.1.0",
				TagName:    "v0.1.0",
				Target:     "master",
				Draft:      false,
				Prerelease: false,
				Body:       "release description",
			},
			expectedRelease: &Release{
				ID:         11111111,
				Name:       "0.1.0",
				TagName:    "v0.1.0",
				Target:     "master",
				Draft:      false,
				Prerelease: false,
				Body:       "release description",
				Assets: []Asset{
					{
						ID:          1111,
						Name:        "app-darwin-amd64",
						Label:       "",
						State:       "uploaded",
						Size:        1048576,
						ContentType: "application/octet-stream",
						DownloadURL: "https://github.com/username/repo/releases/download/v0.1.0/app-darwin-amd64",
					},
					{
						ID:          2222,
						Name:        "app-linux-amd64",
						Label:       "",
						State:       "uploaded",
						Size:        1048576,
						ContentType: "application/octet-stream",
						DownloadURL: "https://github.com/username/repo/releases/download/v0.1.0/app-linux-amd64",
					},
				},
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
				r.Methods("PATCH").Path("/repos/{owner}/{repo}/releases/{id}").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(tc.mockStatusCode)
					w.Write([]byte(tc.mockBody))
				})

				ts := httptest.NewServer(r)
				defer ts.Close()

				gh.apiAddr = ts.URL
			}

			release, err := gh.EditRelease(tc.ctx, tc.repo, tc.releaseID, tc.input)

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
			mockBody:       releaseBody1,
			token:          "github-token",
			ctx:            context.Background(),
			repo:           "username/repo",
			version:        SemVer{Major: 0, Minor: 1, Patch: 0},
			expectedRelease: &Release{
				ID:         11111111,
				Name:       "0.1.0",
				TagName:    "v0.1.0",
				Target:     "master",
				Draft:      false,
				Prerelease: false,
				Body:       "release description",
				Assets: []Asset{
					{
						ID:          1111,
						Name:        "app-darwin-amd64",
						Label:       "",
						State:       "uploaded",
						Size:        1048576,
						ContentType: "application/octet-stream",
						DownloadURL: "https://github.com/username/repo/releases/download/v0.1.0/app-darwin-amd64",
					},
					{
						ID:          2222,
						Name:        "app-linux-amd64",
						Label:       "",
						State:       "uploaded",
						Size:        1048576,
						ContentType: "application/octet-stream",
						DownloadURL: "https://github.com/username/repo/releases/download/v0.1.0/app-linux-amd64",
					},
				},
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

func TestGithubGetReleases(t *testing.T) {
	tests := []struct {
		name             string
		mockAPI          bool
		mockStatusCode   int
		mockBody         string
		token            string
		ctx              context.Context
		repo             string
		expectedError    string
		expectedReleases []Release
	}{
		{
			name:          "RequestError",
			mockAPI:       false,
			token:         "github-token",
			ctx:           context.Background(),
			repo:          "username/repo",
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
			expectedError:  "GET /repos/username/repo/releases 400",
		},
		{
			name:           "InvalidResponse",
			mockAPI:        true,
			mockStatusCode: 200,
			mockBody:       `{ invalid }`,
			token:          "github-token",
			ctx:            context.Background(),
			repo:           "username/repo",
			expectedError:  "invalid character 'i' looking for beginning of object key string",
		},
		{
			name:           "Success",
			mockAPI:        true,
			mockStatusCode: 200,
			mockBody:       fmt.Sprintf("[%s, %s]", releaseBody1, releaseBody2),
			token:          "github-token",
			ctx:            context.Background(),
			repo:           "username/repo",
			expectedReleases: []Release{
				{
					ID:         11111111,
					Name:       "0.1.0",
					TagName:    "v0.1.0",
					Target:     "master",
					Draft:      false,
					Prerelease: false,
					Body:       "release description",
					Assets: []Asset{
						{
							ID:          1111,
							Name:        "app-darwin-amd64",
							Label:       "",
							State:       "uploaded",
							Size:        1048576,
							ContentType: "application/octet-stream",
							DownloadURL: "https://github.com/username/repo/releases/download/v0.1.0/app-darwin-amd64",
						},
						{
							ID:          2222,
							Name:        "app-linux-amd64",
							Label:       "",
							State:       "uploaded",
							Size:        1048576,
							ContentType: "application/octet-stream",
							DownloadURL: "https://github.com/username/repo/releases/download/v0.1.0/app-linux-amd64",
						},
					},
				},
				{
					ID:         22222222,
					Name:       "0.2.0",
					TagName:    "v0.2.0",
					Target:     "master",
					Draft:      false,
					Prerelease: false,
					Body:       "release description",
					Assets: []Asset{
						{
							ID:          3333,
							Name:        "app-darwin-amd64",
							Label:       "",
							State:       "uploaded",
							Size:        1048576,
							ContentType: "application/octet-stream",
							DownloadURL: "https://github.com/username/repo/releases/download/v0.2.0/app-darwin-amd64",
						},
						{
							ID:          4444,
							Name:        "app-linux-amd64",
							Label:       "",
							State:       "uploaded",
							Size:        1048576,
							ContentType: "application/octet-stream",
							DownloadURL: "https://github.com/username/repo/releases/download/v0.2.0/app-linux-amd64",
						},
					},
				},
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
				r.Methods("GET").Path("/repos/{owner}/{repo}/releases").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(tc.mockStatusCode)
					w.Write([]byte(tc.mockBody))
				})

				ts := httptest.NewServer(r)
				defer ts.Close()

				gh.apiAddr = ts.URL
			}

			releases, err := gh.GetReleases(tc.ctx, tc.repo)

			if tc.expectedError != "" {
				assert.Contains(t, err.Error(), tc.expectedError)
				assert.Nil(t, releases)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedReleases, releases)
			}
		})
	}
}

func TestGithubGetLatestRelease(t *testing.T) {
	tests := []struct {
		name            string
		mockAPI         bool
		mockStatusCode  int
		mockBody        string
		token           string
		ctx             context.Context
		repo            string
		expectedError   string
		expectedRelease *Release
	}{
		{
			name:          "RequestError",
			mockAPI:       false,
			token:         "github-token",
			ctx:           context.Background(),
			repo:          "username/repo",
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
			expectedError:  "GET /repos/username/repo/releases/latest 400",
		},
		{
			name:           "InvalidResponse",
			mockAPI:        true,
			mockStatusCode: 200,
			mockBody:       `{ invalid }`,
			token:          "github-token",
			ctx:            context.Background(),
			repo:           "username/repo",
			expectedError:  "invalid character 'i' looking for beginning of object key string",
		},
		{
			name:           "Success",
			mockAPI:        true,
			mockStatusCode: 200,
			mockBody:       releaseBody2,
			token:          "github-token",
			ctx:            context.Background(),
			repo:           "username/repo",
			expectedRelease: &Release{
				ID:         22222222,
				Name:       "0.2.0",
				TagName:    "v0.2.0",
				Target:     "master",
				Draft:      false,
				Prerelease: false,
				Body:       "release description",
				Assets: []Asset{
					{
						ID:          3333,
						Name:        "app-darwin-amd64",
						Label:       "",
						State:       "uploaded",
						Size:        1048576,
						ContentType: "application/octet-stream",
						DownloadURL: "https://github.com/username/repo/releases/download/v0.2.0/app-darwin-amd64",
					},
					{
						ID:          4444,
						Name:        "app-linux-amd64",
						Label:       "",
						State:       "uploaded",
						Size:        1048576,
						ContentType: "application/octet-stream",
						DownloadURL: "https://github.com/username/repo/releases/download/v0.2.0/app-linux-amd64",
					},
				},
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
				r.Methods("GET").Path("/repos/{owner}/{repo}/releases/latest").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(tc.mockStatusCode)
					w.Write([]byte(tc.mockBody))
				})

				ts := httptest.NewServer(r)
				defer ts.Close()

				gh.apiAddr = ts.URL
			}

			release, err := gh.GetLatestRelease(tc.ctx, tc.repo)

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
		mockUploadAPI             bool
		mockUploadAssetStatusCode int
		mockUploadAssetBody       string
		token                     string
		ctx                       context.Context
		release                   *Release
		assets                    []string
		expectedError             string
	}{
		{
			name:  "AssetNotExist",
			token: "github-token",
			ctx:   context.Background(),
			release: &Release{
				ID:        12345678,
				Name:      "0.1.0",
				UploadURL: "https://uploads.github.com/repos/username/repo/releases/12345678/assets{?name,label}",
			},
			assets:        []string{"./test/nil"},
			expectedError: "open test/nil: no such file or directory",
		},
		{
			name:  "EmptyAsset",
			token: "github-token",
			ctx:   context.Background(),
			release: &Release{
				ID:        12345678,
				Name:      "0.1.0",
				UploadURL: "https://uploads.github.com/repos/username/repo/releases/12345678/assets{?name,label}",
			},
			assets:        []string{"./test/empty"},
			expectedError: "EOF",
		},
		{
			name:  "UploadAssetRequestError",
			token: "github-token",
			ctx:   context.Background(),
			release: &Release{
				ID:        12345678,
				Name:      "0.1.0",
				UploadURL: "",
			},
			assets:        []string{"./test/asset"},
			expectedError: "unsupported protocol scheme",
		},
		{
			name:                      "UploadAssetBadStatusCode",
			mockUploadAPI:             true,
			mockUploadAssetStatusCode: 500,
			mockUploadAssetBody:       "",
			token:                     "github-token",
			ctx:                       context.Background(),
			release: &Release{
				ID:        12345678,
				Name:      "0.1.0",
				UploadURL: "https://uploads.github.com/repos/username/repo/releases/12345678/assets{?name,label}",
			},
			assets:        []string{"./test/asset"},
			expectedError: "POST /repos/username/repo/releases/12345678/assets 500",
		},
		{
			name:                      "Successful",
			mockUploadAPI:             true,
			mockUploadAssetStatusCode: 201,
			mockUploadAssetBody:       `{}`,
			token:                     "github-token",
			ctx:                       context.Background(),
			release: &Release{
				ID:        12345678,
				Name:      "0.1.0",
				UploadURL: "https://uploads.github.com/repos/username/repo/releases/12345678/assets{?name,label}",
			},
			assets:        []string{"./test/asset"},
			expectedError: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := &http.Client{}
			gh := &github{
				client:     client,
				authHeader: "token " + tc.token,
			}

			if tc.mockUploadAPI {
				r := mux.NewRouter()
				r.Methods("POST").Path("/repos/{owner}/{repo}/releases/{id}/assets").Queries("name", "{name}").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(tc.mockUploadAssetStatusCode)
					w.Write([]byte(tc.mockUploadAssetBody))
				})

				ts := httptest.NewServer(r)
				defer ts.Close()

				tc.release.UploadURL = strings.Replace(tc.release.UploadURL, "https://uploads.github.com", ts.URL, 1)
			}

			err := gh.UploadAssets(tc.ctx, tc.release, tc.assets...)

			if tc.expectedError != "" {
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
