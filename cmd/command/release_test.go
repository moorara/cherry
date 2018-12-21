package command

import (
	"errors"
	"testing"

	"github.com/moorara/cherry/internal/semver"
	"github.com/stretchr/testify/assert"
)

func TestNewRelease(t *testing.T) {
	tests := []struct {
		name string
	}{}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.NotNil(t, tc)
		})
	}
}

func TestProcessVersions(t *testing.T) {
	tests := []struct {
		name            string
		manager         *mockManager
		releaseType     releaseType
		expectedError   string
		expectedCurrent semver.SemVer
		expectedNext    semver.SemVer
	}{
		{
			name: "ManagerReadError",
			manager: &mockManager{
				ReadOutError: errors.New("invalid version file"),
			},
			releaseType:   patchRelease,
			expectedError: "invalid version file",
		},
		{
			name: "ManagerUpdateError",
			manager: &mockManager{
				UpdateOutError: errors.New("version file error"),
			},
			releaseType:   patchRelease,
			expectedError: "version file error",
		},
		{
			name: "PatchRelease",
			manager: &mockManager{
				ReadOutSemVer: semver.SemVer{Major: 0, Minor: 1, Patch: 0},
			},
			releaseType:     patchRelease,
			expectedCurrent: semver.SemVer{Major: 0, Minor: 1, Patch: 0},
			expectedNext:    semver.SemVer{Major: 0, Minor: 1, Patch: 1},
		},
		{
			name: "MinorRelease",
			manager: &mockManager{
				ReadOutSemVer: semver.SemVer{Major: 0, Minor: 1, Patch: 0},
			},
			releaseType:     minorRelease,
			expectedCurrent: semver.SemVer{Major: 0, Minor: 2, Patch: 0},
			expectedNext:    semver.SemVer{Major: 0, Minor: 2, Patch: 1},
		},
		{
			name: "MajorRelease",
			manager: &mockManager{
				ReadOutSemVer: semver.SemVer{Major: 0, Minor: 1, Patch: 0},
			},
			releaseType:     majorRelease,
			expectedCurrent: semver.SemVer{Major: 1, Minor: 0, Patch: 0},
			expectedNext:    semver.SemVer{Major: 1, Minor: 0, Patch: 1},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd := &Release{
				Manager: tc.manager,
			}

			current, next, err := cmd.processVersions(tc.releaseType)

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
		name string
	}{}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.NotNil(t, tc)
		})
	}
}

func TestReleaseSynopsis(t *testing.T) {
	cmd := &Release{}
	synopsis := cmd.Synopsis()
	assert.Equal(t, releaseSynopsis, synopsis)
}

func TestReleaseHelp(t *testing.T) {
	cmd := &Release{}
	help := cmd.Help()
	assert.Equal(t, releaseHelp, help)
}

func TestReleaseRun(t *testing.T) {
	tests := []struct {
		name string
	}{}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.NotNil(t, tc)
		})
	}
}
