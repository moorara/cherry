package formula

import (
	"context"
	"errors"
	"testing"

	"github.com/moorara/cherry/internal/v1/spec"

	"github.com/moorara/cherry/internal/service/git"
	"github.com/moorara/cherry/internal/service/semver"
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
					GetRepoOutError: errors.New("cannot read git repo"),
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
					GetRepoOutRepo: &git.Repo{
						Owner: "moorara",
						Name:  "cherry",
					},
					GetBranchOutError: errors.New("cannot read git branch"),
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
					GetRepoOutRepo: &git.Repo{
						Owner: "moorara",
						Name:  "cherry",
					},
					GetBranchOutBranch: &git.Branch{},
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
					GetRepoOutRepo: &git.Repo{
						Owner: "moorara",
						Name:  "cherry",
					},
					GetBranchOutBranch: &git.Branch{
						Name: "master",
					},
					IsCleanOutError: errors.New("cannot read git status"),
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
					GetRepoOutRepo: &git.Repo{
						Owner: "moorara",
						Name:  "cherry",
					},
					GetBranchOutBranch: &git.Branch{
						Name: "master",
					},
					IsCleanOutResult: false,
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
					GetRepoOutRepo: &git.Repo{
						Owner: "moorara",
						Name:  "cherry",
					},
					GetBranchOutBranch: &git.Branch{
						Name: "master",
					},
					IsCleanOutResult: true,
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
		expectedCurrent semver.SemVer
		expectedNext    semver.SemVer
	}{
		{
			name: "ManagerReadError",
			formula: &formula{
				Manager: &mockSemVerManager{
					ReadOutError: errors.New("invalid version file"),
				},
			},
			level:         PatchRelease,
			expectedError: "invalid version file",
		},
		{
			name: "ManagerUpdateError",
			formula: &formula{
				Manager: &mockSemVerManager{
					UpdateOutError: errors.New("version file error"),
				},
			},
			level:         PatchRelease,
			expectedError: "version file error",
		},
		{
			name: "PatchRelease",
			formula: &formula{
				Manager: &mockSemVerManager{
					ReadOutSemVer: semver.SemVer{Major: 0, Minor: 1, Patch: 0},
				},
			},
			level:           PatchRelease,
			expectedCurrent: semver.SemVer{Major: 0, Minor: 1, Patch: 0},
			expectedNext:    semver.SemVer{Major: 0, Minor: 1, Patch: 1},
		},
		{
			name: "MinorRelease",
			formula: &formula{
				Manager: &mockSemVerManager{
					ReadOutSemVer: semver.SemVer{Major: 0, Minor: 1, Patch: 0},
				},
			},
			level:           MinorRelease,
			expectedCurrent: semver.SemVer{Major: 0, Minor: 2, Patch: 0},
			expectedNext:    semver.SemVer{Major: 0, Minor: 2, Patch: 1},
		},
		{
			name: "MajorRelease",
			formula: &formula{
				Manager: &mockSemVerManager{
					ReadOutSemVer: semver.SemVer{Major: 0, Minor: 1, Patch: 0},
				},
			},
			level:           MajorRelease,
			expectedCurrent: semver.SemVer{Major: 1, Minor: 0, Patch: 0},
			expectedNext:    semver.SemVer{Major: 1, Minor: 0, Patch: 1},
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
			name: "DisableBranchProtectionFails",
			formula: &formula{
				GithubToken: "github-token",
				Spec:        &spec.Spec{},
				Ui:          &mockUI{},
				Git: &mockGit{
					GetRepoOutRepo: &git.Repo{
						Owner: "moorara",
						Name:  "cherry",
					},
					GetBranchOutBranch: &git.Branch{
						Name: "master",
					},
					IsCleanOutResult: true,
				},
				Github: &mockGithub{
					BranchProtectionForAdminOutError: errors.New("github branch protection api error"),
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
					GetRepoOutRepo: &git.Repo{
						Owner: "moorara",
						Name:  "cherry",
					},
					GetBranchOutBranch: &git.Branch{
						Name: "master",
					},
					IsCleanOutResult: true,
				},
				Github: &mockGithub{
					BranchProtectionForAdminOutError: nil,
				},
				Manager: &mockSemVerManager{
					ReadOutError: errors.New("invalid version file"),
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
					GetRepoOutRepo: &git.Repo{
						Owner: "moorara",
						Name:  "cherry",
					},
					GetBranchOutBranch: &git.Branch{
						Name: "master",
					},
					IsCleanOutResult: true,
				},
				Github: &mockGithub{
					BranchProtectionForAdminOutError: nil,
				},
				Manager: &mockSemVerManager{
					ReadOutSemVer:  semver.SemVer{Major: 0, Minor: 1, Patch: 0},
					UpdateOutError: nil,
				},
				Changelog: &mockChangelog{
					GenerateOutError: errors.New("changelog generation error"),
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
					GetRepoOutRepo: &git.Repo{
						Owner: "moorara",
						Name:  "cherry",
					},
					GetBranchOutBranch: &git.Branch{
						Name: "master",
					},
					IsCleanOutResult: true,
					CommitOutError:   errors.New("git commit error"),
				},
				Github: &mockGithub{
					BranchProtectionForAdminOutError: nil,
				},
				Manager: &mockSemVerManager{
					ReadOutSemVer:  semver.SemVer{Major: 0, Minor: 1, Patch: 0},
					UpdateOutError: nil,
				},
				Changelog: &mockChangelog{
					GenerateOutResult: "changelog",
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
					GetRepoOutRepo: &git.Repo{
						Owner: "moorara",
						Name:  "cherry",
					},
					GetBranchOutBranch: &git.Branch{
						Name: "master",
					},
					IsCleanOutResult: true,
					CommitOutError:   nil,
					TagOutError:      errors.New("git tag error"),
				},
				Github: &mockGithub{
					BranchProtectionForAdminOutError: nil,
				},
				Manager: &mockSemVerManager{
					ReadOutSemVer:  semver.SemVer{Major: 0, Minor: 1, Patch: 0},
					UpdateOutError: nil,
				},
				Changelog: &mockChangelog{
					GenerateOutResult: "changelog",
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
					GetRepoOutRepo: &git.Repo{
						Owner: "moorara",
						Name:  "cherry",
					},
					GetBranchOutBranch: &git.Branch{
						Name: "master",
					},
					IsCleanOutResult: true,
					CommitOutError:   nil,
					TagOutError:      nil,
					PushOutError:     errors.New("git push error"),
				},
				Github: &mockGithub{
					BranchProtectionForAdminOutError: nil,
				},
				Manager: &mockSemVerManager{
					ReadOutSemVer:  semver.SemVer{Major: 0, Minor: 1, Patch: 0},
					UpdateOutError: nil,
				},
				Changelog: &mockChangelog{
					GenerateOutResult: "changelog",
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
					GetRepoOutRepo: &git.Repo{
						Owner: "moorara",
						Name:  "cherry",
					},
					GetBranchOutBranch: &git.Branch{
						Name: "master",
					},
					IsCleanOutResult: true,
					CommitOutError:   nil,
					TagOutError:      nil,
					PushOutError:     nil,
				},
				Github: &mockGithub{
					BranchProtectionForAdminOutError: nil,
					CreateReleaseOutError:            errors.New("github create release error"),
				},
				Manager: &mockSemVerManager{
					ReadOutSemVer:  semver.SemVer{Major: 0, Minor: 1, Patch: 0},
					UpdateOutError: nil,
				},
				Changelog: &mockChangelog{
					GenerateOutResult: "changelog",
				},
			},
			ctx:           context.Background(),
			level:         PatchRelease,
			comment:       "release description",
			expectedError: "github create release error",
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
