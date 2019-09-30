package step

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

type mockHTTP struct {
	Method       string
	Path         string
	StatusCode   int
	ResponseBody string
}

func createMockHTTPServer(mocks ...mockHTTP) http.Handler {
	r := mux.NewRouter()
	for _, m := range mocks {
		r.Methods(m.Method).Path(m.Path).HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(m.StatusCode)
			w.Write([]byte(m.ResponseBody))
		})
	}

	return r
}

func TestHTTPError(t *testing.T) {
	tests := []struct {
		name          string
		request       *http.Request
		statusCode    int
		body          string
		expectedError string
	}{
		{
			"400",
			&http.Request{
				Method: "GET",
				URL: &url.URL{
					Path: "/",
				},
			},
			http.StatusBadRequest,
			"Invalid request",
			"GET / 400: Invalid request",
		},
		{
			"500",
			&http.Request{
				Method: "POST",
				URL: &url.URL{
					Path: "/",
				},
			},
			http.StatusInternalServerError,
			"Internal error",
			"POST / 500: Internal error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			br := strings.NewReader(tc.body)
			rc := ioutil.NopCloser(br)

			res := &http.Response{
				Request:    tc.request,
				StatusCode: tc.statusCode,
				Body:       rc,
			}

			err := newHTTPError(res)
			assert.Equal(t, tc.request, err.Request)
			assert.Equal(t, tc.statusCode, err.StatusCode)
			assert.Equal(t, tc.body, err.Message)
			assert.Equal(t, tc.expectedError, err.Error())
		})
	}
}

func TestGitHubBranchProtectionDry(t *testing.T) {
	tests := []struct {
		name          string
		mockResponses []mockHTTP
		token         string
		repo          string
		branch        string
		expectedError string
	}{
		{
			name:          "RequestError",
			token:         "github-token",
			repo:          "username/repo",
			branch:        "master",
			expectedError: `Get /repos/username/repo/branches/master/protection/enforce_admins: unsupported protocol scheme ""`,
		},
		{
			name: "BadStatusCode",
			mockResponses: []mockHTTP{
				{"GET", "/repos/{owner}/{repo}/branches/{branch}/protection/enforce_admins", 403, ``},
			},
			token:         "github-token",
			repo:          "username/repo",
			branch:        "master",
			expectedError: `GET /repos/username/repo/branches/master/protection/enforce_admins 403: `,
		},
		{
			name: "Enable",
			mockResponses: []mockHTTP{
				{"GET", "/repos/{owner}/{repo}/branches/{branch}/protection/enforce_admins", 200, `{ "enabled": false }`},
			},
			token:  "github-token",
			repo:   "username/repo",
			branch: "master",
		},
		{
			name: "Disable",
			mockResponses: []mockHTTP{
				{"GET", "/repos/{owner}/{repo}/branches/{branch}/protection/enforce_admins", 200, `{ "enabled": true }`},
			},
			token:  "github-token",
			repo:   "username/repo",
			branch: "master",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := &GitHubBranchProtection{
				Client: &http.Client{},
				Ctx:    context.Background(),
				Token:  tc.token,
				Repo:   tc.repo,
				Branch: tc.branch,
			}

			if len(tc.mockResponses) > 0 {
				h := createMockHTTPServer(tc.mockResponses...)
				ts := httptest.NewServer(h)
				defer ts.Close()

				step.BaseURL = ts.URL
			}

			err := step.Dry()
			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}

func TestGitHubBranchProtectionRun(t *testing.T) {
	tests := []struct {
		name          string
		mockResponses []mockHTTP
		token         string
		repo          string
		branch        string
		enabled       bool
		expectedError string
	}{
		{
			name:          "RequestError",
			token:         "github-token",
			repo:          "username/repo",
			branch:        "master",
			enabled:       true,
			expectedError: `Post /repos/username/repo/branches/master/protection/enforce_admins: unsupported protocol scheme ""`,
		},
		{
			name: "BadStatusCode",
			mockResponses: []mockHTTP{
				{"POST", "/repos/{owner}/{repo}/branches/{branch}/protection/enforce_admins", 403, ``},
			},
			token:         "github-token",
			repo:          "username/repo",
			branch:        "master",
			enabled:       true,
			expectedError: `POST /repos/username/repo/branches/master/protection/enforce_admins 403: `,
		},
		{
			name: "Enable",
			mockResponses: []mockHTTP{
				{"POST", "/repos/{owner}/{repo}/branches/{branch}/protection/enforce_admins", 200, `{ "enabled": true }`},
			},
			token:   "github-token",
			repo:    "username/repo",
			branch:  "master",
			enabled: true,
		},
		{
			name: "Disable",
			mockResponses: []mockHTTP{
				{"DELETE", "/repos/{owner}/{repo}/branches/{branch}/protection/enforce_admins", 204, ``},
			},
			token:   "github-token",
			repo:    "username/repo",
			branch:  "master",
			enabled: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := &GitHubBranchProtection{
				Client:  &http.Client{},
				Ctx:     context.Background(),
				Token:   tc.token,
				Repo:    tc.repo,
				Branch:  tc.branch,
				Enabled: tc.enabled,
			}

			if len(tc.mockResponses) > 0 {
				h := createMockHTTPServer(tc.mockResponses...)
				ts := httptest.NewServer(h)
				defer ts.Close()

				step.BaseURL = ts.URL
			}

			err := step.Run()
			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}

func TestGitHubBranchProtectionRevert(t *testing.T) {
	tests := []struct {
		name          string
		mockResponses []mockHTTP
		token         string
		repo          string
		branch        string
		enabled       bool
		expectedError string
	}{
		{
			name:          "RequestError",
			token:         "github-token",
			repo:          "username/repo",
			branch:        "master",
			enabled:       true,
			expectedError: `Delete /repos/username/repo/branches/master/protection/enforce_admins: unsupported protocol scheme ""`,
		},
		{
			name: "BadStatusCode",
			mockResponses: []mockHTTP{
				{"DELETE", "/repos/{owner}/{repo}/branches/{branch}/protection/enforce_admins", 403, ``},
			},
			token:         "github-token",
			repo:          "username/repo",
			branch:        "master",
			enabled:       true,
			expectedError: `DELETE /repos/username/repo/branches/master/protection/enforce_admins 403: `,
		},
		{
			name: "Enable",
			mockResponses: []mockHTTP{
				{"DELETE", "/repos/{owner}/{repo}/branches/{branch}/protection/enforce_admins", 204, ``},
			},
			token:   "github-token",
			repo:    "username/repo",
			branch:  "master",
			enabled: true,
		},
		{
			name: "Disable",
			mockResponses: []mockHTTP{
				{"POST", "/repos/{owner}/{repo}/branches/{branch}/protection/enforce_admins", 200, `{ "enabled": true }`},
			},
			token:   "github-token",
			repo:    "username/repo",
			branch:  "master",
			enabled: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := &GitHubBranchProtection{
				Client:  &http.Client{},
				Ctx:     context.Background(),
				Token:   tc.token,
				Repo:    tc.repo,
				Branch:  tc.branch,
				Enabled: tc.enabled,
			}

			if len(tc.mockResponses) > 0 {
				h := createMockHTTPServer(tc.mockResponses...)
				ts := httptest.NewServer(h)
				defer ts.Close()

				step.BaseURL = ts.URL
			}

			err := step.Revert()
			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}

func TestGitHubGetLatestReleaseDry(t *testing.T) {
	tests := []struct {
		name          string
		mockResponses []mockHTTP
		token         string
		repo          string
		expectedError string
	}{
		{
			name:          "RequestError",
			token:         "github-token",
			repo:          "username/repo",
			expectedError: `Get /repos/username/repo/releases/latest: unsupported protocol scheme ""`,
		},
		{
			name: "BadStatusCode",
			mockResponses: []mockHTTP{
				{"GET", "/repos/{owner}/{repo}/releases/latest", 403, ``},
			},
			token:         "github-token",
			repo:          "username/repo",
			expectedError: `GET /repos/username/repo/releases/latest 403: `,
		},
		{
			name: "Success",
			mockResponses: []mockHTTP{
				{
					"GET", "/repos/{owner}/{repo}/releases/latest", 200, `{
						"id": 1,
						"tag_name": "v0.1.0",
						"target_commitish": "master",
						"name": "0.1.0",
						"body": "comment",
						"draft": false,
						"prerelease": false
					}`,
				},
			},
			token: "github-token",
			repo:  "username/repo",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := &GitHubGetLatestRelease{
				Client: &http.Client{},
				Ctx:    context.Background(),
				Token:  tc.token,
				Repo:   tc.repo,
			}

			if len(tc.mockResponses) > 0 {
				h := createMockHTTPServer(tc.mockResponses...)
				ts := httptest.NewServer(h)
				defer ts.Close()

				step.BaseURL = ts.URL
			}

			err := step.Dry()
			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}

func TestGitHubGetLatestReleaseRun(t *testing.T) {
	tests := []struct {
		name                  string
		mockResponses         []mockHTTP
		token                 string
		repo                  string
		expectedError         string
		expectedLatestRelease GitHubRelease
	}{
		{
			name:          "RequestError",
			token:         "github-token",
			repo:          "username/repo",
			expectedError: `Get /repos/username/repo/releases/latest: unsupported protocol scheme ""`,
		},
		{
			name: "BadStatusCode",
			mockResponses: []mockHTTP{
				{"GET", "/repos/{owner}/{repo}/releases/latest", 403, ``},
			},
			token:         "github-token",
			repo:          "username/repo",
			expectedError: `GET /repos/username/repo/releases/latest 403: `,
		},
		{
			name: "Success",
			mockResponses: []mockHTTP{
				{
					"GET", "/repos/{owner}/{repo}/releases/latest", 200, `{
						"id": 1,
						"tag_name": "v0.1.0",
						"target_commitish": "master",
						"name": "0.1.0",
						"body": "comment",
						"draft": false,
						"prerelease": false
					}`,
				},
			},
			token: "github-token",
			repo:  "username/repo",
			expectedLatestRelease: GitHubRelease{
				ID:         1,
				Name:       "0.1.0",
				TagName:    "v0.1.0",
				Target:     "master",
				Draft:      false,
				Prerelease: false,
				Body:       "comment",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := &GitHubGetLatestRelease{
				Client: &http.Client{},
				Ctx:    context.Background(),
				Token:  tc.token,
				Repo:   tc.repo,
			}

			if len(tc.mockResponses) > 0 {
				h := createMockHTTPServer(tc.mockResponses...)
				ts := httptest.NewServer(h)
				defer ts.Close()

				step.BaseURL = ts.URL
			}

			err := step.Run()
			if tc.expectedError == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedLatestRelease, step.Result.LatestRelease)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}

func TestGitHubGetLatestReleaseRevert(t *testing.T) {
	tests := []struct {
		name          string
		token         string
		repo          string
		expectedError string
	}{
		{
			name:  "Success",
			token: "github-token",
			repo:  "username/repo",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := &GitHubGetLatestRelease{
				Client: &http.Client{},
				Ctx:    context.Background(),
				Token:  tc.token,
				Repo:   tc.repo,
			}

			err := step.Revert()
			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}

func TestGitHubCreateReleaseDry(t *testing.T) {
	tests := []struct {
		name          string
		mockResponses []mockHTTP
		token         string
		repo          string
		expectedError string
	}{
		{
			name:          "RequestError",
			token:         "github-token",
			repo:          "username/repo",
			expectedError: `Get /repos/username/repo/releases/latest: unsupported protocol scheme ""`,
		},
		{
			name: "BadStatusCode",
			mockResponses: []mockHTTP{
				{"GET", "/repos/{owner}/{repo}/releases/latest", 403, ``},
			},
			token:         "github-token",
			repo:          "username/repo",
			expectedError: `GET /repos/username/repo/releases/latest 403: `,
		},
		{
			name: "Success",
			mockResponses: []mockHTTP{
				{
					"GET", "/repos/{owner}/{repo}/releases/latest", 200, `{
						"id": 1,
						"tag_name": "v0.1.0",
						"target_commitish": "master",
						"name": "0.1.0",
						"body": "comment",
						"draft": false,
						"prerelease": false
					}`,
				},
			},
			token: "github-token",
			repo:  "username/repo",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := &GitHubCreateRelease{
				Client: &http.Client{},
				Ctx:    context.Background(),
				Token:  tc.token,
				Repo:   tc.repo,
			}

			if len(tc.mockResponses) > 0 {
				h := createMockHTTPServer(tc.mockResponses...)
				ts := httptest.NewServer(h)
				defer ts.Close()

				step.BaseURL = ts.URL
			}

			err := step.Dry()
			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}

func TestGitHubCreateReleaseRun(t *testing.T) {
	tests := []struct {
		name            string
		mockResponses   []mockHTTP
		token           string
		repo            string
		releaseData     GitHubReleaseData
		expectedError   string
		expectedRelease GitHubRelease
	}{
		{
			name:  "RequestError",
			token: "github-token",
			repo:  "username/repo",
			releaseData: GitHubReleaseData{
				Name:       "0.2.0",
				TagName:    "v0.2.0",
				Target:     "master",
				Draft:      true,
				Prerelease: false,
			},
			expectedError: `Post /repos/username/repo/releases: unsupported protocol scheme ""`,
		},
		{
			name: "BadStatusCode",
			mockResponses: []mockHTTP{
				{"POST", "/repos/{owner}/{repo}/releases", 403, ``},
			},
			token: "github-token",
			repo:  "username/repo",
			releaseData: GitHubReleaseData{
				Name:       "0.2.0",
				TagName:    "v0.2.0",
				Target:     "master",
				Draft:      true,
				Prerelease: false,
			},
			expectedError: `POST /repos/username/repo/releases 403: `,
		},
		{
			name: "Success",
			mockResponses: []mockHTTP{
				{
					"POST", "/repos/{owner}/{repo}/releases", 201, `{
						"id": 2,
						"tag_name": "v0.2.0",
						"target_commitish": "master",
						"name": "0.2.0",
						"body": "",
						"draft": true,
						"prerelease": false
					}`,
				},
			},
			token: "github-token",
			repo:  "username/repo",
			releaseData: GitHubReleaseData{
				Name:       "0.2.0",
				TagName:    "v0.2.0",
				Target:     "master",
				Draft:      true,
				Prerelease: false,
			},
			expectedRelease: GitHubRelease{
				ID:         2,
				Name:       "0.2.0",
				TagName:    "v0.2.0",
				Target:     "master",
				Draft:      true,
				Prerelease: false,
				Body:       "",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := &GitHubCreateRelease{
				Client:      &http.Client{},
				Ctx:         context.Background(),
				Token:       tc.token,
				Repo:        tc.repo,
				ReleaseData: tc.releaseData,
			}

			if len(tc.mockResponses) > 0 {
				h := createMockHTTPServer(tc.mockResponses...)
				ts := httptest.NewServer(h)
				defer ts.Close()

				step.BaseURL = ts.URL
			}

			err := step.Run()
			if tc.expectedError == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedRelease, step.Result.Release)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}

func TestGitHubCreateReleaseRevert(t *testing.T) {
	tests := []struct {
		name          string
		mockResponses []mockHTTP
		token         string
		repo          string
		release       GitHubRelease
		expectedError string
	}{
		{
			name:  "RequestError",
			token: "github-token",
			repo:  "username/repo",
			release: GitHubRelease{
				ID: 2,
			},
			expectedError: `Delete /repos/username/repo/releases/2: unsupported protocol scheme ""`,
		},
		{
			name: "BadStatusCode",
			mockResponses: []mockHTTP{
				{"DELETE", "/repos/{owner}/{repo}/releases/{id}", 403, ``},
			},
			token: "github-token",
			repo:  "username/repo",
			release: GitHubRelease{
				ID: 2,
			},
			expectedError: `DELETE /repos/username/repo/releases/2 403: `,
		},
		{
			name: "Success",
			mockResponses: []mockHTTP{
				{"DELETE", "/repos/{owner}/{repo}/releases/{id}", 204, ``},
			},
			token: "github-token",
			repo:  "username/repo",
			release: GitHubRelease{
				ID: 2,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := &GitHubCreateRelease{
				Client: &http.Client{},
				Ctx:    context.Background(),
				Token:  tc.token,
				Repo:   tc.repo,
			}

			step.Result.Release = tc.release

			if len(tc.mockResponses) > 0 {
				h := createMockHTTPServer(tc.mockResponses...)
				ts := httptest.NewServer(h)
				defer ts.Close()

				step.BaseURL = ts.URL
			}

			err := step.Revert()
			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}

func TestGitHubEditReleaseDry(t *testing.T) {
	tests := []struct {
		name          string
		mockResponses []mockHTTP
		token         string
		repo          string
		releaseID     int
		expectedError string
	}{
		{
			name:          "RequestError",
			token:         "github-token",
			repo:          "username/repo",
			releaseID:     2,
			expectedError: `Get /repos/username/repo/releases/2: unsupported protocol scheme ""`,
		},
		{
			name: "BadStatusCode",
			mockResponses: []mockHTTP{
				{"GET", "/repos/{owner}/{repo}/releases/{id}", 403, ``},
			},
			token:         "github-token",
			repo:          "username/repo",
			releaseID:     2,
			expectedError: `GET /repos/username/repo/releases/2 403: `,
		},
		{
			name: "Success",
			mockResponses: []mockHTTP{
				{
					"GET", "/repos/{owner}/{repo}/releases/{id}", 200, `{
						"id": 2,
						"tag_name": "v0.2.0",
						"target_commitish": "master",
						"name": "0.2.0",
						"body": "",
						"draft": true,
						"prerelease": false
					}`,
				},
			},
			token:     "github-token",
			repo:      "username/repo",
			releaseID: 2,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := &GitHubEditRelease{
				Client:    &http.Client{},
				Ctx:       context.Background(),
				Token:     tc.token,
				Repo:      tc.repo,
				ReleaseID: tc.releaseID,
			}

			if len(tc.mockResponses) > 0 {
				h := createMockHTTPServer(tc.mockResponses...)
				ts := httptest.NewServer(h)
				defer ts.Close()

				step.BaseURL = ts.URL
			}

			err := step.Dry()
			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}

func TestGitHubEditReleaseRun(t *testing.T) {
	tests := []struct {
		name            string
		mockResponses   []mockHTTP
		token           string
		repo            string
		releaseID       int
		releaseData     GitHubReleaseData
		expectedError   string
		expectedRelease GitHubRelease
	}{
		{
			name:      "RequestError",
			token:     "github-token",
			repo:      "username/repo",
			releaseID: 2,
			releaseData: GitHubReleaseData{
				Draft: false,
				Body:  "comment",
			},
			expectedError: `Patch /repos/username/repo/releases/2: unsupported protocol scheme ""`,
		},
		{
			name: "BadStatusCode",
			mockResponses: []mockHTTP{
				{"PATCH", "/repos/{owner}/{repo}/releases/{id}", 403, ``},
			},
			token:         "github-token",
			repo:          "username/repo",
			releaseID:     2,
			expectedError: `PATCH /repos/username/repo/releases/2 403: `,
		},
		{
			name: "Success",
			mockResponses: []mockHTTP{
				{
					"PATCH", "/repos/{owner}/{repo}/releases/{id}", 200, `{
						"id": 2,
						"tag_name": "v0.2.0",
						"target_commitish": "master",
						"name": "0.2.0",
						"body": "comment",
						"draft": false,
						"prerelease": false
					}`,
				},
			},
			token:     "github-token",
			repo:      "username/repo",
			releaseID: 2,
			expectedRelease: GitHubRelease{
				ID:         2,
				Name:       "0.2.0",
				TagName:    "v0.2.0",
				Target:     "master",
				Draft:      false,
				Prerelease: false,
				Body:       "comment",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := &GitHubEditRelease{
				Client:      &http.Client{},
				Ctx:         context.Background(),
				Token:       tc.token,
				Repo:        tc.repo,
				ReleaseID:   tc.releaseID,
				ReleaseData: tc.releaseData,
			}

			if len(tc.mockResponses) > 0 {
				h := createMockHTTPServer(tc.mockResponses...)
				ts := httptest.NewServer(h)
				defer ts.Close()

				step.BaseURL = ts.URL
			}

			err := step.Run()
			if tc.expectedError == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedRelease, step.Result.Release)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}

func TestGitHubEditReleaseRevert(t *testing.T) {
	tests := []struct {
		name          string
		mockResponses []mockHTTP
		token         string
		repo          string
		releaseID     int
		expectedError string
	}{
		{
			name:          "RequestError",
			token:         "github-token",
			repo:          "username/repo",
			releaseID:     2,
			expectedError: ``,
		},
		{
			name:          "BadStatusCode",
			mockResponses: []mockHTTP{},
			token:         "github-token",
			repo:          "username/repo",
			releaseID:     2,
			expectedError: ``,
		},
		{
			name:          "Success",
			mockResponses: []mockHTTP{},
			token:         "github-token",
			repo:          "username/repo",
			releaseID:     2,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := &GitHubEditRelease{
				Client:    &http.Client{},
				Ctx:       context.Background(),
				Token:     tc.token,
				Repo:      tc.repo,
				ReleaseID: tc.releaseID,
			}

			if len(tc.mockResponses) > 0 {
				h := createMockHTTPServer(tc.mockResponses...)
				ts := httptest.NewServer(h)
				defer ts.Close()

				step.BaseURL = ts.URL
			}

			err := step.Revert()
			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}

func TestGitHubUploadAssetsDry(t *testing.T) {
	tests := []struct {
		name          string
		mockResponses []mockHTTP
		token         string
		repo          string
		releaseID     int
		expectedError string
	}{
		{
			name:          "RequestError",
			token:         "github-token",
			repo:          "username/repo",
			releaseID:     2,
			expectedError: `Get /repos/username/repo/releases/2/assets: unsupported protocol scheme ""`,
		},
		{
			name: "BadStatusCode",
			mockResponses: []mockHTTP{
				{"GET", "/repos/{owner}/{repo}/releases/{id}/assets", 403, ``},
			},
			token:         "github-token",
			repo:          "username/repo",
			expectedError: `GET /repos/username/repo/releases/0/assets 403: `,
		},
		{
			name: "Success",
			mockResponses: []mockHTTP{
				{
					"GET", "/repos/{owner}/{repo}/releases/{id}/assets", 200, `[]`,
				},
			},
			token: "github-token",
			repo:  "username/repo",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := &GitHubUploadAssets{
				Client:    &http.Client{},
				Ctx:       context.Background(),
				Token:     tc.token,
				Repo:      tc.repo,
				ReleaseID: tc.releaseID,
			}

			if len(tc.mockResponses) > 0 {
				h := createMockHTTPServer(tc.mockResponses...)
				ts := httptest.NewServer(h)
				defer ts.Close()

				step.BaseURL = ts.URL
			}

			err := step.Dry()
			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}

func TestGitHubUploadAssetsRun(t *testing.T) {
	tests := []struct {
		name               string
		mockResponses      []mockHTTP
		token              string
		repo               string
		releaseID          int
		releaseUploadURL   string
		assetFiles         []string
		expectedErrorRegex string
		expectedAssets     []GitHubAsset
	}{
		{
			name:               "RequestError",
			token:              "github-token",
			repo:               "username/repo",
			releaseID:          2,
			releaseUploadURL:   "/repos/username/repo/releases/2/assets{?name,label}",
			assetFiles:         []string{"./test/asset"},
			expectedErrorRegex: `Post /repos/username/repo/releases/2/assets?name=asset: unsupported protocol scheme ""`,
		},
		{
			name: "BadStatusCode",
			mockResponses: []mockHTTP{
				{"POST", "/repos/{owner}/{repo}/releases/{id}/assets", 403, ``},
			},
			token:              "github-token",
			repo:               "username/repo",
			releaseID:          2,
			releaseUploadURL:   "https://uploads.github.com/repos/username/repo/releases/2/assets{?name,label}",
			assetFiles:         []string{"./test/asset"},
			expectedErrorRegex: `POST /repos/username/repo/releases/2/assets 403: `,
		},
		{
			name: "Success",
			mockResponses: []mockHTTP{
				{
					"POST", "/repos/{owner}/{repo}/releases/{id}/assets", 201, `{
						"id": 1,
						"name": "app",
						"label": "",
						"state": "uploaded",
						"size": 1024,
						"content_type": "application/octet-stream"
					}`,
				},
			},
			token:            "github-token",
			repo:             "username/repo",
			releaseID:        2,
			releaseUploadURL: "https://uploads.github.com/repos/username/repo/releases/2/assets{?name,label}",
			assetFiles:       []string{"./test/asset"},
			expectedAssets: []GitHubAsset{
				{
					ID:          1,
					Name:        "app",
					Label:       "",
					State:       "uploaded",
					Size:        1024,
					ContentType: "application/octet-stream",
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := &GitHubUploadAssets{
				Client:           &http.Client{},
				Ctx:              context.Background(),
				Token:            tc.token,
				Repo:             tc.repo,
				ReleaseID:        tc.releaseID,
				ReleaseUploadURL: tc.releaseUploadURL,
				AssetFiles:       tc.assetFiles,
			}

			if len(tc.mockResponses) > 0 {
				h := createMockHTTPServer(tc.mockResponses...)
				ts := httptest.NewServer(h)
				defer ts.Close()

				step.BaseURL = ts.URL
				step.ReleaseUploadURL = strings.Replace(step.ReleaseUploadURL, "https://uploads.github.com", ts.URL, 1)
			}

			err := step.Run()
			if tc.expectedErrorRegex == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedErrorRegex, err.Error())
			}
		})
	}
}

func TestGitHubUploadAssetsRevert(t *testing.T) {
	tests := []struct {
		name          string
		mockResponses []mockHTTP
		token         string
		repo          string
		assets        []GitHubAsset
		expectedError string
	}{
		{
			name:  "RequestError",
			token: "github-token",
			repo:  "username/repo",
			assets: []GitHubAsset{
				{ID: 1},
			},
			expectedError: `Delete /repos/username/repo/releases/assets/1: unsupported protocol scheme ""`,
		},
		{
			name: "BadStatusCode",
			mockResponses: []mockHTTP{
				{"DELETE", "/repos/{owner}/{repo}/releases/assets/{id}", 403, ``},
			},
			token: "github-token",
			repo:  "username/repo",
			assets: []GitHubAsset{
				{ID: 1},
			},
			expectedError: `DELETE /repos/username/repo/releases/assets/1 403: `,
		},
		{
			name: "Success",
			mockResponses: []mockHTTP{
				{"DELETE", "/repos/{owner}/{repo}/releases/assets/{id}", 204, ``},
			},
			token: "github-token",
			repo:  "username/repo",
			assets: []GitHubAsset{
				{ID: 1},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			step := &GitHubUploadAssets{
				Client: &http.Client{},
				Ctx:    context.Background(),
				Token:  tc.token,
				Repo:   tc.repo,
			}

			step.Result.Assets = tc.assets

			if len(tc.mockResponses) > 0 {
				h := createMockHTTPServer(tc.mockResponses...)
				ts := httptest.NewServer(h)
				defer ts.Close()

				step.BaseURL = ts.URL
			}

			err := step.Revert()
			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}
