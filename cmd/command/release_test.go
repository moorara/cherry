package command

import (
	"io/ioutil"
	"os"
	"path/filepath"
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

func TestGetVersions(t *testing.T) {
	tests := []struct {
		name               string
		workDir            string
		versionFileName    string
		versionFileContent string
		releaseType        releaseType
		expectedError      string
		expectedCurrent    semver.SemVer
		expectedNext       semver.SemVer
	}{
		{
			name:               "NoVersionFile",
			workDir:            "",
			versionFileName:    "",
			versionFileContent: "",
			releaseType:        patchRelease,
			expectedError:      "no such file or directory",
		},
		{
			name:               "EmptyVersionFile",
			workDir:            os.TempDir(),
			versionFileName:    "VERSION_EMPTY",
			versionFileContent: "",
			releaseType:        patchRelease,
			expectedError:      "invalid semantic version",
		},
		{
			name:               "InvalidVersionFile",
			workDir:            os.TempDir(),
			versionFileName:    "VERSION_INVALID",
			versionFileContent: "1.0",
			releaseType:        patchRelease,
			expectedError:      "invalid semantic version",
		},
		{
			name:               "PatchRelease",
			workDir:            os.TempDir(),
			versionFileName:    "VERSION_PATCH",
			versionFileContent: "0.1.0-0",
			releaseType:        patchRelease,
			expectedCurrent:    semver.SemVer{Major: 0, Minor: 1, Patch: 0},
			expectedNext:       semver.SemVer{Major: 0, Minor: 1, Patch: 1},
		},
		{
			name:               "MinorRelease",
			workDir:            os.TempDir(),
			versionFileName:    "VERSION_MINOR",
			versionFileContent: "0.1.0-0",
			releaseType:        minorRelease,
			expectedCurrent:    semver.SemVer{Major: 0, Minor: 2, Patch: 0},
			expectedNext:       semver.SemVer{Major: 0, Minor: 2, Patch: 1},
		},
		{
			name:               "MajorRelease",
			workDir:            os.TempDir(),
			versionFileName:    "VERSION_MAJOR",
			versionFileContent: "0.1.0-0",
			releaseType:        majorRelease,
			expectedCurrent:    semver.SemVer{Major: 1, Minor: 0, Patch: 0},
			expectedNext:       semver.SemVer{Major: 1, Minor: 0, Patch: 1},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.workDir != "" && tc.versionFileName != "" {
				versionFilePath := filepath.Join(tc.workDir, tc.versionFileName)
				err := ioutil.WriteFile(versionFilePath, []byte(tc.versionFileContent), 0644)
				assert.NoError(t, err)
				defer os.Remove(versionFilePath)
			}

			cmd := &Release{
				ui:          &mockUI{},
				workDir:     tc.workDir,
				versionFile: tc.versionFileName,
			}

			current, next, err := cmd.getVersions(tc.releaseType)

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
