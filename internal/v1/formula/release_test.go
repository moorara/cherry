package formula

import (
	"errors"
	"testing"

	"github.com/moorara/cherry/internal/service/semver"
	"github.com/stretchr/testify/assert"
)

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

func TestProcessVersions(t *testing.T) {
	tests := []struct {
		name            string
		manager         *mockManager
		level           ReleaseLevel
		expectedError   string
		expectedCurrent semver.SemVer
		expectedNext    semver.SemVer
	}{
		{
			name: "ManagerReadError",
			manager: &mockManager{
				ReadOutError: errors.New("invalid version file"),
			},
			level:         PatchRelease,
			expectedError: "invalid version file",
		},
		{
			name: "ManagerUpdateError",
			manager: &mockManager{
				UpdateOutError: errors.New("version file error"),
			},
			level:         PatchRelease,
			expectedError: "version file error",
		},
		{
			name: "PatchRelease",
			manager: &mockManager{
				ReadOutSemVer: semver.SemVer{Major: 0, Minor: 1, Patch: 0},
			},
			level:           PatchRelease,
			expectedCurrent: semver.SemVer{Major: 0, Minor: 1, Patch: 0},
			expectedNext:    semver.SemVer{Major: 0, Minor: 1, Patch: 1},
		},
		{
			name: "MinorRelease",
			manager: &mockManager{
				ReadOutSemVer: semver.SemVer{Major: 0, Minor: 1, Patch: 0},
			},
			level:           MinorRelease,
			expectedCurrent: semver.SemVer{Major: 0, Minor: 2, Patch: 0},
			expectedNext:    semver.SemVer{Major: 0, Minor: 2, Patch: 1},
		},
		{
			name: "MajorRelease",
			manager: &mockManager{
				ReadOutSemVer: semver.SemVer{Major: 0, Minor: 1, Patch: 0},
			},
			level:           MajorRelease,
			expectedCurrent: semver.SemVer{Major: 1, Minor: 0, Patch: 0},
			expectedNext:    semver.SemVer{Major: 1, Minor: 0, Patch: 1},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := &release{
				Manager: tc.manager,
			}

			current, next, err := r.processVersions(tc.level)

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
