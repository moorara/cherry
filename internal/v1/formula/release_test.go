package formula

import (
	"context"
	"errors"
	"testing"

	"github.com/moorara/cherry/internal/service"
	"github.com/moorara/cherry/internal/v1/spec"
	"github.com/stretchr/testify/assert"
)

func TestEnsure(t *testing.T) {
	tests := []struct {
		name           string
		formula        *formula
		expectedError  string
		expectedRepo   string
		expectedBranch string
	}{
		{
			name: "NoGitHubToken",
			formula: &formula{
				githubToken: "",
				spec:        &spec.Spec{},
				ui:          &mockUI{},
			},
			expectedError: "github token is not set",
		},
		{
			name: "GitRepoError",
			formula: &formula{
				githubToken: "github-token",
				spec:        &spec.Spec{},
				ui:          &mockUI{},
				git: &mockGit{
					GetRepoMocks: []GetRepoMock{
						{
							OutError: errors.New("cannot read git repo"),
						},
					},
				},
			},
			expectedError: "cannot read git repo",
		},
		{
			name: "GetBranchError",
			formula: &formula{
				githubToken: "github-token",
				spec:        &spec.Spec{},
				ui:          &mockUI{},
				git: &mockGit{
					GetRepoMocks: []GetRepoMock{
						{
							OutRepo: &service.Repo{
								Owner: "moorara",
								Name:  "cherry",
							},
						},
					},
					GetBranchMocks: []GetBranchMock{
						{
							OutError: errors.New("cannot read git branch"),
						},
					},
				},
			},
			expectedError: "cannot read git branch",
		},
		{
			name: "NotOnMasterBranch",
			formula: &formula{
				githubToken: "github-token",
				spec:        &spec.Spec{},
				ui:          &mockUI{},
				git: &mockGit{
					GetRepoMocks: []GetRepoMock{
						{
							OutRepo: &service.Repo{
								Owner: "moorara",
								Name:  "cherry",
							},
						},
					},
					GetBranchMocks: []GetBranchMock{
						{
							OutBranch: &service.Branch{
								Name: "develop",
							},
						},
					},
				},
			},
			expectedError: "release has to be run on master branch",
		},
		{
			name: "IsCleanFails",
			formula: &formula{
				githubToken: "github-token",
				spec:        &spec.Spec{},
				ui:          &mockUI{},
				git: &mockGit{
					GetRepoMocks: []GetRepoMock{
						{
							OutRepo: &service.Repo{
								Owner: "moorara",
								Name:  "cherry",
							},
						},
					},
					GetBranchMocks: []GetBranchMock{
						{
							OutBranch: &service.Branch{
								Name: "master",
							},
						},
					},
					IsCleanMocks: []IsCleanMock{
						{
							OutError: errors.New("cannot read git status"),
						},
					},
				},
			},
			expectedError: "cannot read git status",
		},
		{
			name: "RepoNotClean",
			formula: &formula{
				githubToken: "github-token",
				spec:        &spec.Spec{},
				ui:          &mockUI{},
				git: &mockGit{
					GetRepoMocks: []GetRepoMock{
						{
							OutRepo: &service.Repo{
								Owner: "moorara",
								Name:  "cherry",
							},
						},
					},
					GetBranchMocks: []GetBranchMock{
						{
							OutBranch: &service.Branch{
								Name: "master",
							},
						},
					},
					IsCleanMocks: []IsCleanMock{
						{
							OutResult: false,
						},
					},
				},
			},
			expectedError: "working directory is not clean",
		},
		{
			name: "GitPullFails",
			formula: &formula{
				githubToken: "github-token",
				spec:        &spec.Spec{},
				ui:          &mockUI{},
				git: &mockGit{
					GetRepoMocks: []GetRepoMock{
						{
							OutRepo: &service.Repo{
								Owner: "moorara",
								Name:  "cherry",
							},
						},
					},
					GetBranchMocks: []GetBranchMock{
						{
							OutBranch: &service.Branch{
								Name: "master",
							},
						},
					},
					IsCleanMocks: []IsCleanMock{
						{
							OutResult: true,
						},
					},
					PullMocks: []PullMock{
						{
							OutError: errors.New("git pull error"),
						},
					},
				},
			},
			expectedError: "git pull error",
		},
		{
			name: "Success",
			formula: &formula{
				githubToken: "github-token",
				spec:        &spec.Spec{},
				ui:          &mockUI{},
				git: &mockGit{
					GetRepoMocks: []GetRepoMock{
						{
							OutRepo: &service.Repo{
								Owner: "moorara",
								Name:  "cherry",
							},
						},
					},
					GetBranchMocks: []GetBranchMock{
						{
							OutBranch: &service.Branch{
								Name: "master",
							},
						},
					},
					IsCleanMocks: []IsCleanMock{
						{
							OutResult: true,
						},
					},
					PullMocks: []PullMock{
						{
							OutError: nil,
						},
					},
				},
			},
			expectedRepo:   "moorara/cherry",
			expectedBranch: "master",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo, branch, err := tc.formula.ensure()

			if tc.expectedError != "" {
				assert.Contains(t, err.Error(), tc.expectedError)
				assert.Empty(t, repo)
				assert.Empty(t, branch)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedRepo, repo)
				assert.Equal(t, tc.expectedBranch, branch)
			}
		})
	}
}

func TestVersions(t *testing.T) {
	tests := []struct {
		name            string
		formula         *formula
		level           ReleaseLevel
		expectedError   string
		expectedCurrent service.SemVer
		expectedNext    service.SemVer
	}{
		{
			name: "ManagerReadError",
			formula: &formula{
				vmanager: &mockVersionManager{
					ReadMocks: []ReadMock{
						{
							OutError: errors.New("invalid version file"),
						},
					},
				},
			},
			level:         PatchRelease,
			expectedError: "invalid version file",
		},
		{
			name: "ManagerUpdateError",
			formula: &formula{
				vmanager: &mockVersionManager{
					ReadMocks: []ReadMock{
						{
							OutSemVer: service.SemVer{Major: 0, Minor: 1, Patch: 0},
						},
					},
					UpdateMocks: []UpdateMock{
						{
							OutError: errors.New("version file error"),
						},
					},
				},
			},
			level:         PatchRelease,
			expectedError: "version file error",
		},
		{
			name: "PatchRelease",
			formula: &formula{
				vmanager: &mockVersionManager{
					ReadMocks: []ReadMock{
						{
							OutSemVer: service.SemVer{Major: 0, Minor: 1, Patch: 0},
						},
					},
					UpdateMocks: []UpdateMock{
						{
							OutError: nil,
						},
					},
				},
			},
			level:           PatchRelease,
			expectedCurrent: service.SemVer{Major: 0, Minor: 1, Patch: 0},
			expectedNext:    service.SemVer{Major: 0, Minor: 1, Patch: 1},
		},
		{
			name: "MinorRelease",
			formula: &formula{
				vmanager: &mockVersionManager{
					ReadMocks: []ReadMock{
						{
							OutSemVer: service.SemVer{Major: 0, Minor: 1, Patch: 0},
						},
					},
					UpdateMocks: []UpdateMock{
						{
							OutError: nil,
						},
					},
				},
			},
			level:           MinorRelease,
			expectedCurrent: service.SemVer{Major: 0, Minor: 2, Patch: 0},
			expectedNext:    service.SemVer{Major: 0, Minor: 2, Patch: 1},
		},
		{
			name: "MajorRelease",
			formula: &formula{
				vmanager: &mockVersionManager{
					ReadMocks: []ReadMock{
						{
							OutSemVer: service.SemVer{Major: 0, Minor: 1, Patch: 0},
						},
					},
					UpdateMocks: []UpdateMock{
						{
							OutError: nil,
						},
					},
				},
			},
			level:           MajorRelease,
			expectedCurrent: service.SemVer{Major: 1, Minor: 0, Patch: 0},
			expectedNext:    service.SemVer{Major: 1, Minor: 0, Patch: 1},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			current, next, err := tc.formula.versions(tc.level)

			if tc.expectedError != "" {
				assert.Contains(t, err.Error(), tc.expectedError)
				assert.Empty(t, current)
				assert.Empty(t, next)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedCurrent, current)
				assert.Equal(t, tc.expectedNext, next)
			}
		})
	}
}

func TestRelease(t *testing.T) {
	tests := []struct {
		name          string
		formula       *formula
		ctx           context.Context
		level         ReleaseLevel
		comment       string
		expectedError string
	}{
		{
			name: "EnsureFails",
			formula: &formula{
				githubToken: "",
				spec:        &spec.Spec{},
				ui:          &mockUI{},
			},
			ctx:           context.Background(),
			level:         PatchRelease,
			comment:       "release description",
			expectedError: "github token is not set",
		},
		{
			name: "VersionsFails",
			formula: &formula{
				githubToken: "github-token",
				spec:        &spec.Spec{},
				ui:          &mockUI{},
				git: &mockGit{
					GetRepoMocks: []GetRepoMock{
						{
							OutRepo: &service.Repo{
								Owner: "moorara",
								Name:  "cherry",
							},
						},
					},
					GetBranchMocks: []GetBranchMock{
						{
							OutBranch: &service.Branch{
								Name: "master",
							},
						},
					},
					IsCleanMocks: []IsCleanMock{
						{
							OutResult: true,
						},
					},
					PullMocks: []PullMock{
						{
							OutError: nil,
						},
					},
				},
				vmanager: &mockVersionManager{
					ReadMocks: []ReadMock{
						{
							OutError: errors.New("invalid version file"),
						},
					},
				},
			},
			ctx:           context.Background(),
			level:         PatchRelease,
			comment:       "release description",
			expectedError: "invalid version file",
		},
		{
			name: "CreateReleaseFails",
			formula: &formula{
				githubToken: "github-token",
				spec:        &spec.Spec{},
				ui:          &mockUI{},
				git: &mockGit{
					GetRepoMocks: []GetRepoMock{
						{
							OutRepo: &service.Repo{
								Owner: "moorara",
								Name:  "cherry",
							},
						},
					},
					GetBranchMocks: []GetBranchMock{
						{
							OutBranch: &service.Branch{
								Name: "master",
							},
						},
					},
					IsCleanMocks: []IsCleanMock{
						{
							OutResult: true,
						},
					},
					PullMocks: []PullMock{
						{
							OutError: nil,
						},
					},
				},
				github: &mockGithub{
					CreateReleaseMocks: []CreateReleaseMock{
						{
							OutError: errors.New("github create release error"),
						},
					},
				},
				vmanager: &mockVersionManager{
					ReadMocks: []ReadMock{
						{
							OutSemVer: service.SemVer{Major: 0, Minor: 1, Patch: 0},
						},
					},
					UpdateMocks: []UpdateMock{
						{
							OutError: nil,
						},
					},
				},
			},
			ctx:           context.Background(),
			level:         PatchRelease,
			comment:       "release description",
			expectedError: "github create release error",
		},
		{
			name: "ChangelogGenerateFails",
			formula: &formula{
				githubToken: "github-token",
				spec:        &spec.Spec{},
				ui:          &mockUI{},
				git: &mockGit{
					GetRepoMocks: []GetRepoMock{
						{
							OutRepo: &service.Repo{
								Owner: "moorara",
								Name:  "cherry",
							},
						},
					},
					GetBranchMocks: []GetBranchMock{
						{
							OutBranch: &service.Branch{
								Name: "master",
							},
						},
					},
					IsCleanMocks: []IsCleanMock{
						{
							OutResult: true,
						},
					},
					PullMocks: []PullMock{
						{
							OutError: nil,
						},
					},
				},
				github: &mockGithub{
					CreateReleaseMocks: []CreateReleaseMock{
						{
							OutRelease: &service.Release{
								ID:         12345678,
								Name:       "0.1.0",
								TagName:    "v0.1.0",
								Target:     "master",
								Draft:      true,
								Prerelease: false,
								Body:       "",
							},
						},
					},
				},
				changelog: &mockChangelog{
					GenerateMocks: []GenerateMock{
						{
							OutError: errors.New("changelog generate error"),
						},
					},
				},
				vmanager: &mockVersionManager{
					ReadMocks: []ReadMock{
						{
							OutSemVer: service.SemVer{Major: 0, Minor: 1, Patch: 0},
						},
					},
					UpdateMocks: []UpdateMock{
						{
							OutError: nil,
						},
					},
				},
			},
			ctx:           context.Background(),
			level:         PatchRelease,
			comment:       "release description",
			expectedError: "changelog generate error",
		},
		{
			name: "GitCommitFails",
			formula: &formula{
				githubToken: "github-token",
				spec:        &spec.Spec{},
				ui:          &mockUI{},
				git: &mockGit{
					GetRepoMocks: []GetRepoMock{
						{
							OutRepo: &service.Repo{
								Owner: "moorara",
								Name:  "cherry",
							},
						},
					},
					GetBranchMocks: []GetBranchMock{
						{
							OutBranch: &service.Branch{
								Name: "master",
							},
						},
					},
					IsCleanMocks: []IsCleanMock{
						{
							OutResult: true,
						},
					},
					PullMocks: []PullMock{
						{
							OutError: nil,
						},
					},
					CommitMocks: []CommitMock{
						{
							OutError: errors.New("git commit error"),
						},
					},
				},
				github: &mockGithub{
					CreateReleaseMocks: []CreateReleaseMock{
						{
							OutRelease: &service.Release{
								ID:         12345678,
								Name:       "0.1.0",
								TagName:    "v0.1.0",
								Target:     "master",
								Draft:      true,
								Prerelease: false,
								Body:       "",
							},
						},
					},
				},
				changelog: &mockChangelog{
					FilenameMocks: []FilenameMock{
						{
							OutResult: "CHANGELOG.md",
						},
					},
					GenerateMocks: []GenerateMock{
						{
							OutResult: "changelog text",
						},
					},
				},
				vmanager: &mockVersionManager{
					ReadMocks: []ReadMock{
						{
							OutSemVer: service.SemVer{Major: 0, Minor: 1, Patch: 0},
						},
					},
					UpdateMocks: []UpdateMock{
						{
							OutError: nil,
						},
					},
				},
			},
			ctx:           context.Background(),
			level:         PatchRelease,
			comment:       "release description",
			expectedError: "git commit error",
		},
		{
			name: "GitTagFails",
			formula: &formula{
				githubToken: "github-token",
				spec:        &spec.Spec{},
				ui:          &mockUI{},
				git: &mockGit{
					GetRepoMocks: []GetRepoMock{
						{
							OutRepo: &service.Repo{
								Owner: "moorara",
								Name:  "cherry",
							},
						},
					},
					GetBranchMocks: []GetBranchMock{
						{
							OutBranch: &service.Branch{
								Name: "master",
							},
						},
					},
					IsCleanMocks: []IsCleanMock{
						{
							OutResult: true,
						},
					},
					PullMocks: []PullMock{
						{
							OutError: nil,
						},
					},
					CommitMocks: []CommitMock{
						{
							OutError: nil,
						},
					},
					TagMocks: []TagMock{
						{
							OutError: errors.New("git tag error"),
						},
					},
				},
				github: &mockGithub{
					CreateReleaseMocks: []CreateReleaseMock{
						{
							OutRelease: &service.Release{
								ID:         12345678,
								Name:       "0.1.0",
								TagName:    "v0.1.0",
								Target:     "master",
								Draft:      true,
								Prerelease: false,
								Body:       "",
							},
						},
					},
				},
				changelog: &mockChangelog{
					FilenameMocks: []FilenameMock{
						{
							OutResult: "CHANGELOG.md",
						},
					},
					GenerateMocks: []GenerateMock{
						{
							OutResult: "changelog text",
						},
					},
				},
				vmanager: &mockVersionManager{
					ReadMocks: []ReadMock{
						{
							OutSemVer: service.SemVer{Major: 0, Minor: 1, Patch: 0},
						},
					},
					UpdateMocks: []UpdateMock{
						{
							OutError: nil,
						},
					},
				},
			},
			ctx:           context.Background(),
			level:         PatchRelease,
			comment:       "release description",
			expectedError: "git tag error",
		},
		{
			name: "CrossCompileFails",
			formula: &formula{
				githubToken: "github-token",
				spec: &spec.Spec{
					VersionFile: "VERSION",
					Release: spec.Release{
						Build: true,
					},
				},
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
					GetBranchMocks: []GetBranchMock{
						{
							OutBranch: &service.Branch{
								Name: "master",
							},
						},
					},
					IsCleanMocks: []IsCleanMock{
						{
							OutResult: true,
						},
					},
					PullMocks: []PullMock{
						{
							OutError: nil,
						},
					},
					CommitMocks: []CommitMock{
						{
							OutError: nil,
						},
					},
					TagMocks: []TagMock{
						{
							OutError: nil,
						},
					},
				},
				github: &mockGithub{
					CreateReleaseMocks: []CreateReleaseMock{
						{
							OutRelease: &service.Release{
								ID:         12345678,
								Name:       "0.1.0",
								TagName:    "v0.1.0",
								Target:     "master",
								Draft:      true,
								Prerelease: false,
								Body:       "",
							},
						},
					},
				},
				changelog: &mockChangelog{
					FilenameMocks: []FilenameMock{
						{
							OutResult: "CHANGELOG.md",
						},
					},
					GenerateMocks: []GenerateMock{
						{
							OutResult: "changelog text",
						},
					},
				},
				vmanager: &mockVersionManager{
					ReadMocks: []ReadMock{
						{
							OutSemVer: service.SemVer{Major: 0, Minor: 1, Patch: 0},
						},
					},
					UpdateMocks: []UpdateMock{
						{
							OutError: nil,
						},
					},
				},
			},
			ctx:           context.Background(),
			level:         PatchRelease,
			comment:       "release description",
			expectedError: "no such file or directory",
		},
		{
			name: "NextVersionUpdateFails",
			formula: &formula{
				githubToken: "github-token",
				spec:        &spec.Spec{},
				ui:          &mockUI{},
				git: &mockGit{
					GetRepoMocks: []GetRepoMock{
						{
							OutRepo: &service.Repo{
								Owner: "moorara",
								Name:  "cherry",
							},
						},
					},
					GetBranchMocks: []GetBranchMock{
						{
							OutBranch: &service.Branch{
								Name: "master",
							},
						},
					},
					IsCleanMocks: []IsCleanMock{
						{
							OutResult: true,
						},
					},
					PullMocks: []PullMock{
						{
							OutError: nil,
						},
					},
					CommitMocks: []CommitMock{
						{
							OutError: nil,
						},
					},
					TagMocks: []TagMock{
						{
							OutError: nil,
						},
					},
				},
				github: &mockGithub{
					CreateReleaseMocks: []CreateReleaseMock{
						{
							OutRelease: &service.Release{
								ID:         12345678,
								Name:       "0.1.0",
								TagName:    "v0.1.0",
								Target:     "master",
								Draft:      true,
								Prerelease: false,
								Body:       "",
							},
						},
					},
				},
				changelog: &mockChangelog{
					FilenameMocks: []FilenameMock{
						{
							OutResult: "CHANGELOG.md",
						},
					},
					GenerateMocks: []GenerateMock{
						{
							OutResult: "changelog text",
						},
					},
				},
				vmanager: &mockVersionManager{
					ReadMocks: []ReadMock{
						{
							OutSemVer: service.SemVer{Major: 0, Minor: 1, Patch: 0},
						},
					},
					UpdateMocks: []UpdateMock{
						{
							OutError: nil,
						},
						{
							OutError: errors.New("version manager update error"),
						},
					},
				},
			},
			ctx:           context.Background(),
			level:         PatchRelease,
			comment:       "release description",
			expectedError: "version manager update error",
		},
		{
			name: "NextVersionCommitFails",
			formula: &formula{
				githubToken: "github-token",
				spec:        &spec.Spec{},
				ui:          &mockUI{},
				git: &mockGit{
					GetRepoMocks: []GetRepoMock{
						{
							OutRepo: &service.Repo{
								Owner: "moorara",
								Name:  "cherry",
							},
						},
					},
					GetBranchMocks: []GetBranchMock{
						{
							OutBranch: &service.Branch{
								Name: "master",
							},
						},
					},
					IsCleanMocks: []IsCleanMock{
						{
							OutResult: true,
						},
					},
					PullMocks: []PullMock{
						{
							OutError: nil,
						},
					},
					CommitMocks: []CommitMock{
						{
							OutError: nil,
						},
						{
							OutError: errors.New("git commit error"),
						},
					},
					TagMocks: []TagMock{
						{
							OutError: nil,
						},
					},
				},
				github: &mockGithub{
					CreateReleaseMocks: []CreateReleaseMock{
						{
							OutRelease: &service.Release{
								ID:         12345678,
								Name:       "0.1.0",
								TagName:    "v0.1.0",
								Target:     "master",
								Draft:      true,
								Prerelease: false,
								Body:       "",
							},
						},
					},
				},
				changelog: &mockChangelog{
					FilenameMocks: []FilenameMock{
						{
							OutResult: "CHANGELOG.md",
						},
					},
					GenerateMocks: []GenerateMock{
						{
							OutResult: "changelog text",
						},
					},
				},
				vmanager: &mockVersionManager{
					ReadMocks: []ReadMock{
						{
							OutSemVer: service.SemVer{Major: 0, Minor: 1, Patch: 0},
						},
					},
					UpdateMocks: []UpdateMock{
						{
							OutError: nil,
						},
						{
							OutError: nil,
						},
					},
				},
			},
			ctx:           context.Background(),
			level:         PatchRelease,
			comment:       "release description",
			expectedError: "git commit error",
		},
		{
			name: "DisableBranchProtectionFails",
			formula: &formula{
				githubToken: "github-token",
				spec:        &spec.Spec{},
				ui:          &mockUI{},
				git: &mockGit{
					GetRepoMocks: []GetRepoMock{
						{
							OutRepo: &service.Repo{
								Owner: "moorara",
								Name:  "cherry",
							},
						},
					},
					GetBranchMocks: []GetBranchMock{
						{
							OutBranch: &service.Branch{
								Name: "master",
							},
						},
					},
					IsCleanMocks: []IsCleanMock{
						{
							OutResult: true,
						},
					},
					PullMocks: []PullMock{
						{
							OutError: nil,
						},
					},
					CommitMocks: []CommitMock{
						{
							OutError: nil,
						},
						{
							OutError: nil,
						},
					},
					TagMocks: []TagMock{
						{
							OutError: nil,
						},
					},
				},
				github: &mockGithub{
					CreateReleaseMocks: []CreateReleaseMock{
						{
							OutRelease: &service.Release{
								ID:         12345678,
								Name:       "0.1.0",
								TagName:    "v0.1.0",
								Target:     "master",
								Draft:      true,
								Prerelease: false,
								Body:       "",
							},
						},
					},
					BranchProtectionForAdminMocks: []BranchProtectionForAdminMock{
						{
							OutError: errors.New("github disable branch protection error"),
						},
					},
				},
				changelog: &mockChangelog{
					FilenameMocks: []FilenameMock{
						{
							OutResult: "CHANGELOG.md",
						},
					},
					GenerateMocks: []GenerateMock{
						{
							OutResult: "changelog text",
						},
					},
				},
				vmanager: &mockVersionManager{
					ReadMocks: []ReadMock{
						{
							OutSemVer: service.SemVer{Major: 0, Minor: 1, Patch: 0},
						},
					},
					UpdateMocks: []UpdateMock{
						{
							OutError: nil,
						},
						{
							OutError: nil,
						},
					},
				},
			},
			ctx:           context.Background(),
			level:         PatchRelease,
			comment:       "release description",
			expectedError: "github disable branch protection error",
		},
		{
			name: "GitPushFails",
			formula: &formula{
				githubToken: "github-token",
				spec:        &spec.Spec{},
				ui:          &mockUI{},
				git: &mockGit{
					GetRepoMocks: []GetRepoMock{
						{
							OutRepo: &service.Repo{
								Owner: "moorara",
								Name:  "cherry",
							},
						},
					},
					GetBranchMocks: []GetBranchMock{
						{
							OutBranch: &service.Branch{
								Name: "master",
							},
						},
					},
					IsCleanMocks: []IsCleanMock{
						{
							OutResult: true,
						},
					},
					PullMocks: []PullMock{
						{
							OutError: nil,
						},
					},
					CommitMocks: []CommitMock{
						{
							OutError: nil,
						},
						{
							OutError: nil,
						},
					},
					TagMocks: []TagMock{
						{
							OutError: nil,
						},
					},
					PushMocks: []PushMock{
						{
							OutError: errors.New("git push error"),
						},
					},
				},
				github: &mockGithub{
					CreateReleaseMocks: []CreateReleaseMock{
						{
							OutRelease: &service.Release{
								ID:         12345678,
								Name:       "0.1.0",
								TagName:    "v0.1.0",
								Target:     "master",
								Draft:      true,
								Prerelease: false,
								Body:       "",
							},
						},
					},
					BranchProtectionForAdminMocks: []BranchProtectionForAdminMock{
						{
							OutError: nil,
						},
						{
							OutError: nil,
						},
					},
				},
				changelog: &mockChangelog{
					FilenameMocks: []FilenameMock{
						{
							OutResult: "CHANGELOG.md",
						},
					},
					GenerateMocks: []GenerateMock{
						{
							OutResult: "changelog text",
						},
					},
				},
				vmanager: &mockVersionManager{
					ReadMocks: []ReadMock{
						{
							OutSemVer: service.SemVer{Major: 0, Minor: 1, Patch: 0},
						},
					},
					UpdateMocks: []UpdateMock{
						{
							OutError: nil,
						},
						{
							OutError: nil,
						},
					},
				},
			},
			ctx:           context.Background(),
			level:         PatchRelease,
			comment:       "release description",
			expectedError: "git push error",
		},
		{
			name: "EditReleaseFails",
			formula: &formula{
				githubToken: "github-token",
				spec:        &spec.Spec{},
				ui:          &mockUI{},
				git: &mockGit{
					GetRepoMocks: []GetRepoMock{
						{
							OutRepo: &service.Repo{
								Owner: "moorara",
								Name:  "cherry",
							},
						},
					},
					GetBranchMocks: []GetBranchMock{
						{
							OutBranch: &service.Branch{
								Name: "master",
							},
						},
					},
					IsCleanMocks: []IsCleanMock{
						{
							OutResult: true,
						},
					},
					PullMocks: []PullMock{
						{
							OutError: nil,
						},
					},
					CommitMocks: []CommitMock{
						{
							OutError: nil,
						},
						{
							OutError: nil,
						},
					},
					TagMocks: []TagMock{
						{
							OutError: nil,
						},
					},
					PushMocks: []PushMock{
						{
							OutError: nil,
						},
					},
				},
				github: &mockGithub{
					CreateReleaseMocks: []CreateReleaseMock{
						{
							OutRelease: &service.Release{
								ID:         12345678,
								Name:       "0.1.0",
								TagName:    "v0.1.0",
								Target:     "master",
								Draft:      true,
								Prerelease: false,
								Body:       "",
							},
						},
					},
					BranchProtectionForAdminMocks: []BranchProtectionForAdminMock{
						{
							OutError: nil,
						},
						{
							OutError: nil,
						},
					},
					EditReleaseMocks: []EditReleaseMock{
						{
							OutError: errors.New("github edit release error"),
						},
					},
				},
				changelog: &mockChangelog{
					FilenameMocks: []FilenameMock{
						{
							OutResult: "CHANGELOG.md",
						},
					},
					GenerateMocks: []GenerateMock{
						{
							OutResult: "changelog text",
						},
					},
				},
				vmanager: &mockVersionManager{
					ReadMocks: []ReadMock{
						{
							OutSemVer: service.SemVer{Major: 0, Minor: 1, Patch: 0},
						},
					},
					UpdateMocks: []UpdateMock{
						{
							OutError: nil,
						},
						{
							OutError: nil,
						},
					},
				},
			},
			ctx:           context.Background(),
			level:         PatchRelease,
			comment:       "release description",
			expectedError: "github edit release error",
		},
		{
			name: "EnableBranchProtectionFails",
			formula: &formula{
				githubToken: "github-token",
				spec:        &spec.Spec{},
				ui:          &mockUI{},
				git: &mockGit{
					GetRepoMocks: []GetRepoMock{
						{
							OutRepo: &service.Repo{
								Owner: "moorara",
								Name:  "cherry",
							},
						},
					},
					GetBranchMocks: []GetBranchMock{
						{
							OutBranch: &service.Branch{
								Name: "master",
							},
						},
					},
					IsCleanMocks: []IsCleanMock{
						{
							OutResult: true,
						},
					},
					PullMocks: []PullMock{
						{
							OutError: nil,
						},
					},
					CommitMocks: []CommitMock{
						{
							OutError: nil,
						},
						{
							OutError: nil,
						},
					},
					TagMocks: []TagMock{
						{
							OutError: nil,
						},
					},
					PushMocks: []PushMock{
						{
							OutError: nil,
						},
					},
				},
				github: &mockGithub{
					CreateReleaseMocks: []CreateReleaseMock{
						{
							OutRelease: &service.Release{
								ID:         12345678,
								Name:       "0.1.0",
								TagName:    "v0.1.0",
								Target:     "master",
								Draft:      true,
								Prerelease: false,
								Body:       "",
							},
						},
					},
					BranchProtectionForAdminMocks: []BranchProtectionForAdminMock{
						{
							OutError: nil,
						},
						{
							OutError: errors.New("github enable branch protection error"),
						},
					},
					EditReleaseMocks: []EditReleaseMock{
						{
							OutRelease: &service.Release{
								ID:         12345678,
								Name:       "0.1.0",
								TagName:    "v0.1.0",
								Target:     "master",
								Draft:      false,
								Prerelease: false,
								Body:       "release description\n\nchangelog text",
							},
						},
					},
				},
				changelog: &mockChangelog{
					FilenameMocks: []FilenameMock{
						{
							OutResult: "CHANGELOG.md",
						},
					},
					GenerateMocks: []GenerateMock{
						{
							OutResult: "changelog text",
						},
					},
				},
				vmanager: &mockVersionManager{
					ReadMocks: []ReadMock{
						{
							OutSemVer: service.SemVer{Major: 0, Minor: 1, Patch: 0},
						},
					},
					UpdateMocks: []UpdateMock{
						{
							OutError: nil,
						},
						{
							OutError: nil,
						},
					},
				},
			},
			ctx:           context.Background(),
			level:         PatchRelease,
			comment:       "release description",
			expectedError: "",
		},
		{
			name: "Success",
			formula: &formula{
				githubToken: "github-token",
				spec:        &spec.Spec{},
				ui:          &mockUI{},
				git: &mockGit{
					GetRepoMocks: []GetRepoMock{
						{
							OutRepo: &service.Repo{
								Owner: "moorara",
								Name:  "cherry",
							},
						},
					},
					GetBranchMocks: []GetBranchMock{
						{
							OutBranch: &service.Branch{
								Name: "master",
							},
						},
					},
					IsCleanMocks: []IsCleanMock{
						{
							OutResult: true,
						},
					},
					PullMocks: []PullMock{
						{
							OutError: nil,
						},
					},
					CommitMocks: []CommitMock{
						{
							OutError: nil,
						},
						{
							OutError: nil,
						},
					},
					TagMocks: []TagMock{
						{
							OutError: nil,
						},
					},
					PushMocks: []PushMock{
						{
							OutError: nil,
						},
					},
				},
				github: &mockGithub{
					CreateReleaseMocks: []CreateReleaseMock{
						{
							OutRelease: &service.Release{
								ID:         12345678,
								Name:       "0.1.0",
								TagName:    "v0.1.0",
								Target:     "master",
								Draft:      true,
								Prerelease: false,
								Body:       "",
							},
						},
					},
					BranchProtectionForAdminMocks: []BranchProtectionForAdminMock{
						{
							OutError: nil,
						},
						{
							OutError: nil,
						},
					},
					EditReleaseMocks: []EditReleaseMock{
						{
							OutRelease: &service.Release{
								ID:         12345678,
								Name:       "0.1.0",
								TagName:    "v0.1.0",
								Target:     "master",
								Draft:      false,
								Prerelease: false,
								Body:       "release description\n\nchangelog text",
							},
						},
					},
				},
				changelog: &mockChangelog{
					FilenameMocks: []FilenameMock{
						{
							OutResult: "CHANGELOG.md",
						},
					},
					GenerateMocks: []GenerateMock{
						{
							OutResult: "changelog text",
						},
					},
				},
				vmanager: &mockVersionManager{
					ReadMocks: []ReadMock{
						{
							OutSemVer: service.SemVer{Major: 0, Minor: 1, Patch: 0},
						},
					},
					UpdateMocks: []UpdateMock{
						{
							OutError: nil,
						},
						{
							OutError: nil,
						},
					},
				},
			},
			ctx:           context.Background(),
			level:         PatchRelease,
			comment:       "release description",
			expectedError: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.formula.Release(tc.ctx, tc.level, tc.comment)

			if tc.expectedError != "" {
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
