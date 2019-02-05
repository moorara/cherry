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
		PullMockCounter  int
		PullMocks        []PullMock
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

	PullMock struct {
		OutError error
	}
)

func (m *mockGit) IsClean() (bool, error) {
	mock := &m.IsCleanMocks[m.IsCleanCounter]
	m.IsCleanCounter++
	return mock.OutResult, mock.OutError
}

func (m *mockGit) GetRepo() (*service.Repo, error) {
	mock := &m.GetRepoMocks[m.GetRepoCounter]
	m.GetRepoCounter++
	return mock.OutRepo, mock.OutError
}

func (m *mockGit) GetBranch() (*service.Branch, error) {
	mock := &m.GetBranchMocks[m.GetBranchCounter]
	m.GetBranchCounter++
	return mock.OutBranch, mock.OutError
}

func (m *mockGit) GetHEAD() (*service.Commit, error) {
	mock := &m.GetHEADMocks[m.GetHEADCounter]
	m.GetHEADCounter++
	return mock.OutCommit, mock.OutError
}

func (m *mockGit) Commit(message string, files ...string) error {
	mock := &m.CommitMocks[m.CommitCounter]
	m.CommitCounter++
	mock.InMessage = message
	mock.InFiles = files
	return mock.OutError
}

func (m *mockGit) Tag(tag string) error {
	mock := &m.TagMocks[m.TagCounter]
	m.TagCounter++
	mock.InTag = tag
	return mock.OutError
}

func (m *mockGit) Push(withTags bool) error {
	mock := &m.PushMocks[m.PushCounter]
	m.PushCounter++
	mock.InWithTags = withTags
	return mock.OutError
}

func (m *mockGit) Pull() error {
	mock := &m.PullMocks[m.PullMockCounter]
	m.PullMockCounter++
	return mock.OutError
}

type (
	mockGithub struct {
		BranchProtectionForAdminCounter int
		BranchProtectionForAdminMocks   []BranchProtectionForAdminMock
		CreateReleaseCounter            int
		CreateReleaseMocks              []CreateReleaseMock
		EditReleaseCounter              int
		EditReleaseMocks                []EditReleaseMock
		GetReleaseCounter               int
		GetReleaseMocks                 []GetReleaseMock
		GetReleasesCounter              int
		GetReleasesMocks                []GetReleasesMock
		GetLatestReleaseCounter         int
		GetLatestReleaseMocks           []GetLatestReleaseMock
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
		InCtx      context.Context
		InRepo     string
		InInput    service.ReleaseInput
		OutRelease *service.Release
		OutError   error
	}

	EditReleaseMock struct {
		InCtx       context.Context
		InRepo      string
		InReleaseID int
		InInput     service.ReleaseInput
		OutRelease  *service.Release
		OutError    error
	}

	GetReleaseMock struct {
		InCtx      context.Context
		InRepo     string
		InVersion  service.SemVer
		OutRelease *service.Release
		OutError   error
	}

	GetReleasesMock struct {
		InCtx       context.Context
		InRepo      string
		OutReleases []service.Release
		OutError    error
	}

	GetLatestReleaseMock struct {
		InCtx      context.Context
		InRepo     string
		OutRelease *service.Release
		OutError   error
	}

	UploadAssetsMock struct {
		InCtx     context.Context
		InRelease *service.Release
		InAssets  []string
		OutError  error
	}
)

func (m *mockGithub) BranchProtectionForAdmin(ctx context.Context, repo, branch string, enabled bool) error {
	mock := &m.BranchProtectionForAdminMocks[m.BranchProtectionForAdminCounter]
	m.BranchProtectionForAdminCounter++
	mock.InCtx = ctx
	mock.InRepo = repo
	mock.InBranch = branch
	mock.InEnabled = enabled
	return mock.OutError
}

func (m *mockGithub) CreateRelease(ctx context.Context, repo string, input service.ReleaseInput) (*service.Release, error) {
	mock := &m.CreateReleaseMocks[m.CreateReleaseCounter]
	m.CreateReleaseCounter++
	mock.InCtx = ctx
	mock.InRepo = repo
	mock.InInput = input
	return mock.OutRelease, mock.OutError
}

func (m *mockGithub) EditRelease(ctx context.Context, repo string, releaseID int, input service.ReleaseInput) (*service.Release, error) {
	mock := &m.EditReleaseMocks[m.EditReleaseCounter]
	m.EditReleaseCounter++
	mock.InCtx = ctx
	mock.InRepo = repo
	mock.InReleaseID = releaseID
	mock.InInput = input
	return mock.OutRelease, mock.OutError
}

func (m *mockGithub) GetRelease(ctx context.Context, repo string, version service.SemVer) (*service.Release, error) {
	mock := &m.GetReleaseMocks[m.GetReleaseCounter]
	m.GetReleaseCounter++
	mock.InCtx = ctx
	mock.InRepo = repo
	mock.InVersion = version
	return mock.OutRelease, mock.OutError
}

func (m *mockGithub) GetReleases(ctx context.Context, repo string) ([]service.Release, error) {
	mock := &m.GetReleasesMocks[m.GetReleasesCounter]
	m.GetReleasesCounter++
	mock.InCtx = ctx
	mock.InRepo = repo
	return mock.OutReleases, mock.OutError
}

func (m *mockGithub) GetLatestRelease(ctx context.Context, repo string) (*service.Release, error) {
	mock := &m.GetLatestReleaseMocks[m.GetLatestReleaseCounter]
	m.GetLatestReleaseCounter++
	mock.InCtx = ctx
	mock.InRepo = repo
	return mock.OutRelease, mock.OutError
}

func (m *mockGithub) UploadAssets(ctx context.Context, release *service.Release, assets ...string) error {
	mock := &m.UploadAssetsMocks[m.UploadAssetsCounter]
	m.UploadAssetsCounter++
	mock.InCtx = ctx
	mock.InRelease = release
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
		InRepo    string
		InTag     string
		OutResult string
		OutError  error
	}
)

func (m *mockChangelog) Filename() string {
	mock := &m.FilenameMocks[m.FilenameCounter]
	m.FilenameCounter++
	return mock.OutResult
}

func (m *mockChangelog) Generate(ctx context.Context, repo, tag string) (string, error) {
	mock := &m.GenerateMocks[m.GenerateCounter]
	m.GenerateCounter++
	mock.InCtx = ctx
	mock.InRepo = repo
	mock.InTag = tag
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
	mock := &m.ReadMocks[m.ReadCounter]
	m.ReadCounter++
	return mock.OutSemVer, mock.OutError
}

func (m *mockVersionManager) Update(version string) error {
	mock := &m.UpdateMocks[m.UpdateCounter]
	m.UpdateCounter++
	mock.InVersion = version
	return mock.OutError
}
