package formula

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"

	"github.com/gorilla/mux"
	"github.com/moorara/cherry/internal/service"
	"github.com/stretchr/testify/assert"
)

func TestPickRelease(t *testing.T) {
	tests := []struct {
		name            string
		formula         *formula
		releases        []service.Release
		assetName       string
		expectedVersion string
		expectedURL     string
		expectedError   string
	}{
		{
			name:    "Found",
			formula: &formula{},
			releases: []service.Release{
				{
					ID:         11111111,
					Name:       "0.1.0",
					TagName:    "v0.1.0",
					Target:     "master",
					Draft:      false,
					Prerelease: false,
					Body:       "release description",
					Assets: []service.Asset{
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
					Assets: []service.Asset{
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
			assetName:       "app-darwin-amd64",
			expectedVersion: "0.1.0",
			expectedURL:     "https://github.com/username/repo/releases/download/v0.1.0/app-darwin-amd64",
		},
		{
			name:    "NotFound",
			formula: &formula{},
			releases: []service.Release{
				{
					ID:         11111111,
					Name:       "0.1.0",
					TagName:    "v0.1.0",
					Target:     "master",
					Draft:      false,
					Prerelease: false,
					Body:       "release description",
					Assets: []service.Asset{
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
					Assets: []service.Asset{
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
			assetName:     "app-windows-amd64",
			expectedError: "cannot find a release with asset app-windows-amd64",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			version, url, err := tc.formula.pickRelease(tc.releases, tc.assetName)

			if tc.expectedError != "" {
				assert.Contains(t, err.Error(), tc.expectedError)
				assert.Empty(t, version)
				assert.Empty(t, url)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedVersion, version)
				assert.Equal(t, tc.expectedURL, url)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	var mockStatusCode int
	var mockBody string

	r := mux.NewRouter()
	r.Methods("GET").Path("/{owner}/{repo}/releases/download/{tag}/{asset}").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(mockStatusCode)
		w.Write([]byte(mockBody))
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	tests := []struct {
		name           string
		mockStatusCode int
		mockBody       string
		formula        *formula
		ctx            context.Context
		binPath        string
		expectedError  string
	}{
		{
			name: "GitGetRepoFails",
			formula: &formula{
				ui: &mockUI{},
				git: &mockGit{
					GetRepoMocks: []GetRepoMock{
						{
							OutError: errors.New("git repo error"),
						},
					},
				},
			},
			ctx:           context.Background(),
			binPath:       "/dev/null",
			expectedError: "git repo error",
		},
		{
			name: "GithubGetReleases",
			formula: &formula{
				ui: &mockUI{},
				git: &mockGit{
					GetRepoMocks: []GetRepoMock{
						{
							OutRepo: &service.Repo{
								Owner: "moorara",
								Name:  "cherry",
							},
						},
					},
				},
				github: &mockGithub{
					GetReleasesMocks: []GetReleasesMock{
						{
							OutError: errors.New("github get releases error"),
						},
					},
				},
			},
			ctx:           context.Background(),
			binPath:       "/dev/null",
			expectedError: "github get releases error",
		},
		{
			name: "PickReleaseFails",
			formula: &formula{
				ui: &mockUI{},
				git: &mockGit{
					GetRepoMocks: []GetRepoMock{
						{
							OutRepo: &service.Repo{
								Owner: "moorara",
								Name:  "cherry",
							},
						},
					},
				},
				github: &mockGithub{
					GetReleasesMocks: []GetReleasesMock{
						{
							OutReleases: []service.Release{
								{
									ID:      11111111,
									Name:    "0.1.0",
									TagName: "v0.1.0",
									Target:  "master",
									Body:    "release description",
									Assets: []service.Asset{
										{
											ID:          1111,
											Name:        "app",
											State:       "uploaded",
											Size:        1048576,
											ContentType: "application/octet-stream",
											DownloadURL: fmt.Sprintf("%s/username/repo/releases/download/v0.1.0/app", ts.URL),
										},
									},
								},
							},
						},
					},
				},
			},
			ctx:           context.Background(),
			binPath:       "/dev/null",
			expectedError: "cannot find a release with asset cherry",
		},
		{
			name: "OpenFileFails",
			formula: &formula{
				ui: &mockUI{},
				git: &mockGit{
					GetRepoMocks: []GetRepoMock{
						{
							OutRepo: &service.Repo{
								Owner: "moorara",
								Name:  "cherry",
							},
						},
					},
				},
				github: &mockGithub{
					GetReleasesMocks: []GetReleasesMock{
						{
							OutReleases: []service.Release{
								{
									ID:      11111111,
									Name:    "0.1.0",
									TagName: "v0.1.0",
									Target:  "master",
									Body:    "release description",
									Assets: []service.Asset{
										{
											ID:          1111,
											Name:        fmt.Sprintf("cherry-%s-%s", runtime.GOOS, runtime.GOARCH),
											State:       "uploaded",
											Size:        1048576,
											ContentType: "application/octet-stream",
											DownloadURL: fmt.Sprintf("%s/username/repo/releases/download/v0.1.0/cherry-%s-%s", ts.URL, runtime.GOOS, runtime.GOARCH),
										},
									},
								},
							},
						},
					},
				},
			},
			ctx:           context.Background(),
			binPath:       "not_exist",
			expectedError: "no such file or directory",
		},
		{
			name: "NewRequestFails",
			formula: &formula{
				ui: &mockUI{},
				git: &mockGit{
					GetRepoMocks: []GetRepoMock{
						{
							OutRepo: &service.Repo{
								Owner: "moorara",
								Name:  "cherry",
							},
						},
					},
				},
				github: &mockGithub{
					GetReleasesMocks: []GetReleasesMock{
						{
							OutReleases: []service.Release{
								{
									ID:      11111111,
									Name:    "0.1.0",
									TagName: "v0.1.0",
									Target:  "master",
									Body:    "release description",
									Assets: []service.Asset{
										{
											ID:          1111,
											Name:        fmt.Sprintf("cherry-%s-%s", runtime.GOOS, runtime.GOARCH),
											State:       "uploaded",
											Size:        1048576,
											ContentType: "application/octet-stream",
											DownloadURL: ":invalid",
										},
									},
								},
							},
						},
					},
				},
			},
			ctx:           context.Background(),
			binPath:       "/dev/null",
			expectedError: "missing protocol scheme",
		},
		{
			name: "HTTPDoFails",
			formula: &formula{
				ui: &mockUI{},
				git: &mockGit{
					GetRepoMocks: []GetRepoMock{
						{
							OutRepo: &service.Repo{
								Owner: "moorara",
								Name:  "cherry",
							},
						},
					},
				},
				github: &mockGithub{
					GetReleasesMocks: []GetReleasesMock{
						{
							OutReleases: []service.Release{
								{
									ID:      11111111,
									Name:    "0.1.0",
									TagName: "v0.1.0",
									Target:  "master",
									Body:    "release description",
									Assets: []service.Asset{
										{
											ID:          1111,
											Name:        fmt.Sprintf("cherry-%s-%s", runtime.GOOS, runtime.GOARCH),
											State:       "uploaded",
											Size:        1048576,
											ContentType: "application/octet-stream",
											DownloadURL: "unsupported",
										},
									},
								},
							},
						},
					},
				},
			},
			ctx:           context.Background(),
			binPath:       "/dev/null",
			expectedError: "unsupported protocol scheme",
		},
		{
			name:           "BadStatusCode",
			mockStatusCode: 500,
			mockBody:       "server internal error",
			formula: &formula{
				ui: &mockUI{},
				git: &mockGit{
					GetRepoMocks: []GetRepoMock{
						{
							OutRepo: &service.Repo{
								Owner: "moorara",
								Name:  "cherry",
							},
						},
					},
				},
				github: &mockGithub{
					GetReleasesMocks: []GetReleasesMock{
						{
							OutReleases: []service.Release{
								{
									ID:      11111111,
									Name:    "0.1.0",
									TagName: "v0.1.0",
									Target:  "master",
									Body:    "release description",
									Assets: []service.Asset{
										{
											ID:          1111,
											Name:        fmt.Sprintf("cherry-%s-%s", runtime.GOOS, runtime.GOARCH),
											State:       "uploaded",
											Size:        1048576,
											ContentType: "application/octet-stream",
											DownloadURL: fmt.Sprintf("%s/username/repo/releases/download/v0.1.0/cherry-%s-%s", ts.URL, runtime.GOOS, runtime.GOARCH),
										},
									},
								},
							},
						},
					},
				},
			},
			ctx:           context.Background(),
			binPath:       "/dev/null",
			expectedError: "server internal error",
		},
		{
			name:           "Success",
			mockStatusCode: 200,
			mockBody:       "asset content",
			formula: &formula{
				ui: &mockUI{},
				git: &mockGit{
					GetRepoMocks: []GetRepoMock{
						{
							OutRepo: &service.Repo{
								Owner: "moorara",
								Name:  "cherry",
							},
						},
					},
				},
				github: &mockGithub{
					GetReleasesMocks: []GetReleasesMock{
						{
							OutReleases: []service.Release{
								{
									ID:      11111111,
									Name:    "0.1.0",
									TagName: "v0.1.0",
									Target:  "master",
									Body:    "release description",
									Assets: []service.Asset{
										{
											ID:          1111,
											Name:        fmt.Sprintf("cherry-%s-%s", runtime.GOOS, runtime.GOARCH),
											State:       "uploaded",
											Size:        1048576,
											ContentType: "application/octet-stream",
											DownloadURL: fmt.Sprintf("%s/username/repo/releases/download/v0.1.0/cherry-%s-%s", ts.URL, runtime.GOOS, runtime.GOARCH),
										},
									},
								},
							},
						},
					},
				},
			},
			ctx:           context.Background(),
			binPath:       "/dev/null",
			expectedError: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockStatusCode = tc.mockStatusCode
			mockBody = tc.mockBody

			err := tc.formula.Update(tc.ctx, tc.binPath)

			if tc.expectedError != "" {
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
