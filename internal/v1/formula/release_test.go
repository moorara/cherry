package formula

import (
	"context"
	"errors"
	"testing"

	"github.com/moorara/cherry/internal/service"
	"github.com/moorara/cherry/internal/v1/spec"
	"github.com/stretchr/testify/assert"
)

func TestPrecheck(t *testing.T) {
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
				GithubToken: "",
				Spec:        &spec.Spec{},
				Ui:          &mockUI{},
			},
			expectedError: "github token is not set",
		},
		{
			name: "GitRepoError",
			formula: &formula{
				GithubToken: "github-token",
				Spec:        &spec.Spec{},
				Ui:          &mockUI{},
				Git: &mockGit{
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
				GithubToken: "github-token",
				Spec:        &spec.Spec{},
				Ui:          &mockUI{},
				Git: &mockGit{
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
				GithubToken: "github-token",
				Spec:        &spec.Spec{},
				Ui:          &mockUI{},
				Git: &mockGit{
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
				GithubToken: "github-token",
				Spec:        &spec.Spec{},
				Ui:          &mockUI{},
				Git: &mockGit{
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
				GithubToken: "github-token",
				Spec:        &spec.Spec{},
				Ui:          &mockUI{},
				Git: &mockGit{
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
			name: "Success",
			formula: &formula{
				GithubToken: "github-token",
				Spec:        &spec.Spec{},
				Ui:          &mockUI{},
				Git: &mockGit{
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
				},
			},
			expectedRepo:   "moorara/cherry",
			expectedBranch: "master",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo, branch, err := tc.formula.precheck()

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
				VersionManager: &mockVersionManager{
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
				VersionManager: &mockVersionManager{
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
				VersionManager: &mockVersionManager{
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
				VersionManager: &mockVersionManager{
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
				VersionManager: &mockVersionManager{
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
			name: "PrecheckFails",
			formula: &formula{
				GithubToken: "",
				Spec:        &spec.Spec{},
				Ui:          &mockUI{},
			},
			ctx:           context.Background(),
			level:         PatchRelease,
			comment:       "release description",
			expectedError: "github token is not set",
		},
		{
			name: "GitPullFails",
			formula: &formula{
				GithubToken: "github-token",
				Spec:        &spec.Spec{},
				Ui:          &mockUI{},
				Git: &mockGit{
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
			ctx:           context.Background(),
			level:         PatchRelease,
			comment:       "release description",
			expectedError: "git pull error",
		},
		{
			name: "DisableBranchProtectionFails",
			formula: &formula{
				GithubToken: "github-token",
				Spec:        &spec.Spec{},
				Ui:          &mockUI{},
				Git: &mockGit{
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
				Github: &mockGithub{
					BranchProtectionForAdminMocks: []BranchProtectionForAdminMock{
						{
							OutError: errors.New("github branch protection api error"),
						},
					},
				},
			},
			ctx:           context.Background(),
			level:         PatchRelease,
			comment:       "release description",
			expectedError: "github branch protection api error",
		},
		{
			name: "VersionsFails",
			formula: &formula{
				GithubToken: "github-token",
				Spec:        &spec.Spec{},
				Ui:          &mockUI{},
				Git: &mockGit{
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
				Github: &mockGithub{
					BranchProtectionForAdminMocks: []BranchProtectionForAdminMock{
						{
							OutError: nil,
						},
						{
							OutError: nil,
						},
					},
				},
				VersionManager: &mockVersionManager{
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
			name: "ChangelogFails",
			formula: &formula{
				GithubToken: "github-token",
				Spec:        &spec.Spec{},
				Ui:          &mockUI{},
				Git: &mockGit{
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
				Github: &mockGithub{
					BranchProtectionForAdminMocks: []BranchProtectionForAdminMock{
						{
							OutError: nil,
						},
						{
							OutError: nil,
						},
					},
				},
				VersionManager: &mockVersionManager{
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
				Changelog: &mockChangelog{
					GenerateMocks: []GenerateMock{
						{
							OutError: errors.New("changelog generation error"),
						},
					},
				},
			},
			ctx:           context.Background(),
			level:         PatchRelease,
			comment:       "release description",
			expectedError: "changelog generation error",
		},
		{
			name: "GitCommitFails",
			formula: &formula{
				GithubToken: "github-token",
				Spec:        &spec.Spec{},
				Ui:          &mockUI{},
				Git: &mockGit{
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
				Github: &mockGithub{
					BranchProtectionForAdminMocks: []BranchProtectionForAdminMock{
						{
							OutError: nil,
						},
						{
							OutError: nil,
						},
					},
				},
				VersionManager: &mockVersionManager{
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
				Changelog: &mockChangelog{
					FilenameMocks: []FilenameMock{
						{
							OutResult: "CHANGELOG.md",
						},
					},
					GenerateMocks: []GenerateMock{
						{
							OutResult: "changelog",
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
				GithubToken: "github-token",
				Spec:        &spec.Spec{},
				Ui:          &mockUI{},
				Git: &mockGit{
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
				Github: &mockGithub{
					BranchProtectionForAdminMocks: []BranchProtectionForAdminMock{
						{
							OutError: nil,
						},
						{
							OutError: nil,
						},
					},
				},
				VersionManager: &mockVersionManager{
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
				Changelog: &mockChangelog{
					FilenameMocks: []FilenameMock{
						{
							OutResult: "CHANGELOG.md",
						},
					},
					GenerateMocks: []GenerateMock{
						{
							OutResult: "changelog",
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
			name: "GitPushFails",
			formula: &formula{
				GithubToken: "github-token",
				Spec:        &spec.Spec{},
				Ui:          &mockUI{},
				Git: &mockGit{
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
					PushMocks: []PushMock{
						{
							OutError: errors.New("git push error"),
						},
					},
				},
				Github: &mockGithub{
					BranchProtectionForAdminMocks: []BranchProtectionForAdminMock{
						{
							OutError: nil,
						},
						{
							OutError: nil,
						},
					},
				},
				VersionManager: &mockVersionManager{
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
				Changelog: &mockChangelog{
					FilenameMocks: []FilenameMock{
						{
							OutResult: "CHANGELOG.md",
						},
					},
					GenerateMocks: []GenerateMock{
						{
							OutResult: "changelog",
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
			name: "CreateReleaseFails",
			formula: &formula{
				GithubToken: "github-token",
				Spec:        &spec.Spec{},
				Ui:          &mockUI{},
				Git: &mockGit{
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
					PushMocks: []PushMock{
						{
							OutError: nil,
						},
					},
				},
				Github: &mockGithub{
					BranchProtectionForAdminMocks: []BranchProtectionForAdminMock{
						{
							OutError: nil,
						},
						{
							OutError: nil,
						},
					},
					CreateReleaseMocks: []CreateReleaseMock{
						{
							OutError: errors.New("github create release error"),
						},
					},
				},
				VersionManager: &mockVersionManager{
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
				Changelog: &mockChangelog{
					FilenameMocks: []FilenameMock{
						{
							OutResult: "CHANGELOG.md",
						},
					},
					GenerateMocks: []GenerateMock{
						{
							OutResult: "changelog",
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
			name: "NextVersionUpdateFails",
			formula: &formula{
				GithubToken: "github-token",
				Spec:        &spec.Spec{},
				Ui:          &mockUI{},
				Git: &mockGit{
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
					PushMocks: []PushMock{
						{
							OutError: nil,
						},
					},
				},
				Github: &mockGithub{
					BranchProtectionForAdminMocks: []BranchProtectionForAdminMock{
						{
							OutError: nil,
						},
						{
							OutError: nil,
						},
					},
					CreateReleaseMocks: []CreateReleaseMock{
						{
							OutRelease: &service.Release{
								ID:   1,
								Name: "0.1.0",
							},
						},
					},
				},
				VersionManager: &mockVersionManager{
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
				Changelog: &mockChangelog{
					FilenameMocks: []FilenameMock{
						{
							OutResult: "CHANGELOG.md",
						},
					},
					GenerateMocks: []GenerateMock{
						{
							OutResult: "changelog",
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
				GithubToken: "github-token",
				Spec:        &spec.Spec{},
				Ui:          &mockUI{},
				Git: &mockGit{
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
					PushMocks: []PushMock{
						{
							OutError: nil,
						},
					},
				},
				Github: &mockGithub{
					BranchProtectionForAdminMocks: []BranchProtectionForAdminMock{
						{
							OutError: nil,
						},
						{
							OutError: nil,
						},
					},
					CreateReleaseMocks: []CreateReleaseMock{
						{
							OutRelease: &service.Release{
								ID:   1,
								Name: "0.1.0",
							},
						},
					},
				},
				VersionManager: &mockVersionManager{
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
				Changelog: &mockChangelog{
					FilenameMocks: []FilenameMock{
						{
							OutResult: "CHANGELOG.md",
						},
					},
					GenerateMocks: []GenerateMock{
						{
							OutResult: "changelog",
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
			name: "NextVersionPushFails",
			formula: &formula{
				GithubToken: "github-token",
				Spec:        &spec.Spec{},
				Ui:          &mockUI{},
				Git: &mockGit{
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
						{
							OutError: errors.New("git push error"),
						},
					},
				},
				Github: &mockGithub{
					BranchProtectionForAdminMocks: []BranchProtectionForAdminMock{
						{
							OutError: nil,
						},
						{
							OutError: nil,
						},
					},
					CreateReleaseMocks: []CreateReleaseMock{
						{
							OutRelease: &service.Release{
								ID:   1,
								Name: "0.1.0",
							},
						},
					},
				},
				VersionManager: &mockVersionManager{
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
				Changelog: &mockChangelog{
					FilenameMocks: []FilenameMock{
						{
							OutResult: "CHANGELOG.md",
						},
					},
					GenerateMocks: []GenerateMock{
						{
							OutResult: "changelog",
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
			name: "Success",
			formula: &formula{
				GithubToken: "github-token",
				Spec:        &spec.Spec{},
				Ui:          &mockUI{},
				Git: &mockGit{
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
						{
							OutError: nil,
						},
					},
				},
				Github: &mockGithub{
					BranchProtectionForAdminMocks: []BranchProtectionForAdminMock{
						{
							OutError: nil,
						},
						{
							OutError: nil,
						},
					},
					CreateReleaseMocks: []CreateReleaseMock{
						{
							OutRelease: &service.Release{
								ID:   1,
								Name: "0.1.0",
							},
						},
					},
				},
				VersionManager: &mockVersionManager{
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
				Changelog: &mockChangelog{
					FilenameMocks: []FilenameMock{
						{
							OutResult: "CHANGELOG.md",
						},
					},
					GenerateMocks: []GenerateMock{
						{
							OutResult: "changelog",
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
