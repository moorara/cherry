package command

import "github.com/moorara/cherry/internal/semver"

type mockManager struct {
	ReadOutError    error
	ReadOutSemVer   semver.SemVer
	UpdateInVersion string
	UpdateOutError  error
}

func (m *mockManager) Read() (semver.SemVer, error) {
	return m.ReadOutSemVer, m.ReadOutError
}

func (m *mockManager) Update(version string) error {
	m.UpdateInVersion = version
	return m.UpdateOutError
}
