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

type mockGit struct {
	IsCleanOutResult bool
	IsCleanOutError  error

	GetRepoOutRepo  *service.Repo
	GetRepoOutError error

	GetBranchOutBranch *service.Branch
	GetBranchOutError  error

	GetHEADOutCommit *service.Commit
	GetHEADOutError  error

	CommitInMessage string
	CommitInFiles   []string
	CommitOutError  error

	TagInTag    string
	TagOutError error

	PushInWithTags bool
	PushOutError   error
}

func (m *mockGit) IsClean() (bool, error) {
	return m.IsCleanOutResult, m.IsCleanOutError
}

func (m *mockGit) GetRepo() (*service.Repo, error) {
	return m.GetRepoOutRepo, m.GetRepoOutError
}

func (m *mockGit) GetBranch() (*service.Branch, error) {
	return m.GetBranchOutBranch, m.GetBranchOutError
}

func (m *mockGit) GetHEAD() (*service.Commit, error) {
	return m.GetHEADOutCommit, m.GetHEADOutError
}

func (m *mockGit) Commit(message string, files ...string) error {
	m.CommitInMessage = message
	m.CommitInFiles = files
	return m.CommitOutError
}

func (m *mockGit) Tag(tag string) error {
	m.TagInTag = tag
	return m.TagOutError
}

func (m *mockGit) Push(withTags bool) error {
	m.PushInWithTags = withTags
	return m.PushOutError
}

type mockGithub struct {
	BranchProtectionForAdminInCtx     context.Context
	BranchProtectionForAdminInRepo    string
	BranchProtectionForAdminInBranch  string
	BranchProtectionForAdminInEnabled bool
	BranchProtectionForAdminOutError  error

	CreateReleaseInCtx         context.Context
	CreateReleaseInRepo        string
	CreateReleaseInBranch      string
	CreateReleaseInVersion     service.SemVer
	CreateReleaseInDescription string
	CreateReleaseInDraf        bool
	CreateReleaseInPrerelease  bool
	CreateReleaseOutRelease    *service.Release
	CreateReleaseOutError      error

	GetReleaseInCtx      context.Context
	GetReleaseInRepo     string
	GetReleaseInVersion  service.SemVer
	GetReleaseOutRelease *service.Release
	GetReleaseOutError   error

	UploadAssetsInCtx     context.Context
	UploadAssetsInRepo    string
	UploadAssetsInVersion service.SemVer
	UploadAssetsInAssets  []string
	UploadAssetsOutError  error
}

func (m *mockGithub) BranchProtectionForAdmin(ctx context.Context, repo, branch string, enabled bool) error {
	m.BranchProtectionForAdminInCtx = ctx
	m.BranchProtectionForAdminInRepo = repo
	m.BranchProtectionForAdminInBranch = branch
	m.BranchProtectionForAdminInEnabled = enabled
	return m.BranchProtectionForAdminOutError
}

func (m *mockGithub) CreateRelease(ctx context.Context, repo, branch string, version service.SemVer, description string, draf, prerelease bool) (*service.Release, error) {
	m.CreateReleaseInCtx = ctx
	m.CreateReleaseInRepo = repo
	m.CreateReleaseInBranch = branch
	m.CreateReleaseInVersion = version
	m.CreateReleaseInDescription = description
	m.CreateReleaseInDraf = draf
	m.CreateReleaseInPrerelease = prerelease
	return m.CreateReleaseOutRelease, m.CreateReleaseOutError
}

func (m *mockGithub) GetRelease(ctx context.Context, repo string, version service.SemVer) (*service.Release, error) {
	m.GetReleaseInCtx = ctx
	m.GetReleaseInRepo = repo
	m.GetReleaseInVersion = version
	return m.GetReleaseOutRelease, m.GetReleaseOutError
}

func (m *mockGithub) UploadAssets(ctx context.Context, repo string, version service.SemVer, assets []string) error {
	m.UploadAssetsInCtx = ctx
	m.UploadAssetsInRepo = repo
	m.UploadAssetsInVersion = version
	m.UploadAssetsInAssets = assets
	return m.UploadAssetsOutError
}

type mockChangelog struct {
	FilenameOutResult string

	GenerateInCtx     context.Context
	GenerateInGitTag  string
	GenerateOutResult string
	GenerateOutError  error
}

func (m *mockChangelog) Filename() string {
	return m.FilenameOutResult
}

func (m *mockChangelog) Generate(ctx context.Context, gitTag string) (string, error) {
	m.GenerateInCtx = ctx
	m.GenerateInGitTag = gitTag
	return m.GenerateOutResult, m.GenerateOutError
}

type mockVersionManager struct {
	ReadOutSemVer service.SemVer
	ReadOutError  error

	UpdateInVersion string
	UpdateOutError  error
}

func (m *mockVersionManager) Read() (service.SemVer, error) {
	return m.ReadOutSemVer, m.ReadOutError
}

func (m *mockVersionManager) Update(version string) error {
	m.UpdateInVersion = version
	return m.UpdateOutError
}
