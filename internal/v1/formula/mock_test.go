package formula

import (
	"context"

	"github.com/moorara/cherry/internal/service"
)

type mockUI struct {
	OutputInMessage string

	InfoInMessage string

	WarnInMessage string

	ErrorInMessage string

	AskInQuery   string
	AskOutResult string
	AskOutError  error

	AskSecretInQuery   string
	AskSecretOutResult string
	AskSecretOutError  error
}

func (m *mockUI) Output(message string) {
	m.OutputInMessage = message
}

func (m *mockUI) Info(message string) {
	m.InfoInMessage = message
}

func (m *mockUI) Warn(message string) {
	m.WarnInMessage = message
}

func (m *mockUI) Error(message string) {
	m.ErrorInMessage = message
}

func (m *mockUI) Ask(query string) (string, error) {
	m.AskInQuery = query
	return m.AskOutResult, m.AskOutError
}

func (m *mockUI) AskSecret(query string) (string, error) {
	m.AskSecretInQuery = query
	return m.AskSecretOutResult, m.AskSecretOutError
}

type (
	mockGit struct {
		IsCleanCounter   int
		IsCleanMocks     []IsCleanMock
		GetRepoCounter   int
		GetRepoMocks     []GetRepoMock
		GetBranchCounter int
		GetBranchMocks   []GetBranchMock
		GetHEADCounter   int
		GetHEADMocks     []GetHEADMock
		CommitCounter    int
		CommitMocks      []CommitMock
		TagCounter       int
		TagMocks         []TagMock
		PushCounter      int
		PushMocks        []PushMock
	}

	IsCleanMock struct {
		OutResult bool
		OutError  error
	}

	GetRepoMock struct {
		OutRepo  *service.Repo
		OutError error
	}

	GetBranchMock struct {
		OutBranch *service.Branch
		OutError  error
	}

	GetHEADMock struct {
		OutCommit *service.Commit
		OutError  error
	}

	CommitMock struct {
		InMessage string
		InFiles   []string
		OutError  error
	}

	TagMock struct {
		InTag    string
		OutError error
	}

	PushMock struct {
		InWithTags bool
		OutError   error
	}
)

func (m *mockGit) IsClean() (bool, error) {
	i := m.IsCleanCounter
	m.IsCleanCounter++
	mock := m.IsCleanMocks[i]
	return mock.OutResult, mock.OutError
}

func (m *mockGit) GetRepo() (*service.Repo, error) {
	i := m.GetRepoCounter
	m.GetRepoCounter++
	mock := m.GetRepoMocks[i]
	return mock.OutRepo, mock.OutError
}

func (m *mockGit) GetBranch() (*service.Branch, error) {
	i := m.GetBranchCounter
	m.GetBranchCounter++
	mock := m.GetBranchMocks[i]
	return mock.OutBranch, mock.OutError
}

func (m *mockGit) GetHEAD() (*service.Commit, error) {
	i := m.GetHEADCounter
	m.GetHEADCounter++
	mock := m.GetHEADMocks[i]
	return mock.OutCommit, mock.OutError
}

func (m *mockGit) Commit(message string, files ...string) error {
	i := m.CommitCounter
	m.CommitCounter++
	mock := m.CommitMocks[i]
	mock.InMessage = message
	mock.InFiles = files
	return mock.OutError
}

func (m *mockGit) Tag(tag string) error {
	i := m.TagCounter
	m.TagCounter++
	mock := m.TagMocks[i]
	mock.InTag = tag
	return mock.OutError
}

func (m *mockGit) Push(withTags bool) error {
	i := m.PushCounter
	m.PushCounter++
	mock := m.PushMocks[i]
	mock.InWithTags = withTags
	return mock.OutError
}

type (
	mockGithub struct {
		BranchProtectionForAdminCounter int
		BranchProtectionForAdminMocks   []BranchProtectionForAdminMock
		CreateReleaseCounter            int
		CreateReleaseMocks              []CreateReleaseMock
		GetReleaseCounter               int
		GetReleaseMocks                 []GetReleaseMock
		UploadAssetsCounter             int
		UploadAssetsMocks               []UploadAssetsMock
	}

	BranchProtectionForAdminMock struct {
		InCtx     context.Context
		InRepo    string
		InBranch  string
		InEnabled bool
		OutError  error
	}

	CreateReleaseMock struct {
		InCtx         context.Context
		InRepo        string
		InBranch      string
		InVersion     service.SemVer
		InDescription string
		InDraf        bool
		InPrerelease  bool
		OutRelease    *service.Release
		OutError      error
	}

	GetReleaseMock struct {
		InCtx      context.Context
		InRepo     string
		InVersion  service.SemVer
		OutRelease *service.Release
		OutError   error
	}

	UploadAssetsMock struct {
		InCtx     context.Context
		InRepo    string
		InVersion service.SemVer
		InAssets  []string
		OutError  error
	}
)

func (m *mockGithub) BranchProtectionForAdmin(ctx context.Context, repo, branch string, enabled bool) error {
	i := m.BranchProtectionForAdminCounter
	m.BranchProtectionForAdminCounter++
	mock := m.BranchProtectionForAdminMocks[i]
	mock.InCtx = ctx
	mock.InRepo = repo
	mock.InBranch = branch
	mock.InEnabled = enabled
	return mock.OutError
}

func (m *mockGithub) CreateRelease(ctx context.Context, repo, branch string, version service.SemVer, description string, draf, prerelease bool) (*service.Release, error) {
	i := m.CreateReleaseCounter
	m.CreateReleaseCounter++
	mock := m.CreateReleaseMocks[i]
	mock.InCtx = ctx
	mock.InRepo = repo
	mock.InBranch = branch
	mock.InVersion = version
	mock.InDescription = description
	mock.InDraf = draf
	mock.InPrerelease = prerelease
	return mock.OutRelease, mock.OutError
}

func (m *mockGithub) GetRelease(ctx context.Context, repo string, version service.SemVer) (*service.Release, error) {
	i := m.GetReleaseCounter
	m.GetReleaseCounter++
	mock := m.GetReleaseMocks[i]
	mock.InCtx = ctx
	mock.InRepo = repo
	mock.InVersion = version
	return mock.OutRelease, mock.OutError
}

func (m *mockGithub) UploadAssets(ctx context.Context, repo string, version service.SemVer, assets []string) error {
	i := m.UploadAssetsCounter
	m.UploadAssetsCounter++
	mock := m.UploadAssetsMocks[i]
	mock.InCtx = ctx
	mock.InRepo = repo
	mock.InVersion = version
	mock.InAssets = assets
	return mock.OutError
}

type (
	mockChangelog struct {
		FilenameCounter int
		FilenameMocks   []FilenameMock
		GenerateCounter int
		GenerateMocks   []GenerateMock
	}

	FilenameMock struct {
		OutResult string
	}

	GenerateMock struct {
		InCtx     context.Context
		InGitTag  string
		OutResult string
		OutError  error
	}
)

func (m *mockChangelog) Filename() string {
	i := m.FilenameCounter
	m.FilenameCounter++
	mock := m.FilenameMocks[i]
	return mock.OutResult
}

func (m *mockChangelog) Generate(ctx context.Context, gitTag string) (string, error) {
	i := m.GenerateCounter
	m.GenerateCounter++
	mock := m.GenerateMocks[i]
	mock.InCtx = ctx
	mock.InGitTag = gitTag
	return mock.OutResult, mock.OutError
}

type (
	mockVersionManager struct {
		ReadCounter   int
		ReadMocks     []ReadMock
		UpdateCounter int
		UpdateMocks   []UpdateMock
	}

	ReadMock struct {
		OutSemVer service.SemVer
		OutError  error
	}

	UpdateMock struct {
		InVersion string
		OutError  error
	}
)

func (m *mockVersionManager) Read() (service.SemVer, error) {
	i := m.ReadCounter
	m.ReadCounter++
	mock := m.ReadMocks[i]
	return mock.OutSemVer, mock.OutError
}

func (m *mockVersionManager) Update(version string) error {
	i := m.UpdateCounter
	m.UpdateCounter++
	mock := m.UpdateMocks[i]
	mock.InVersion = version
	return mock.OutError
}
