package semver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewManager(t *testing.T) {
	tests := []struct {
		name          string
		file          string
		expectedError string
		expectedType  Manager
	}{
		{
			name:          "Uknown",
			file:          "version.yaml",
			expectedError: "unknown version file",
		},
		{
			name:         "TextManager",
			file:         "VERSION",
			expectedType: &textManager{},
		},
		{
			name:         "JSONManager",
			file:         "package.json",
			expectedType: &jsonManager{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			manager, err := NewManager(tc.file)

			if tc.expectedError != "" {
				assert.Contains(t, err.Error(), tc.expectedError)
				assert.Nil(t, manager)
			} else {
				assert.NoError(t, err)
				assert.IsType(t, tc.expectedType, manager)
			}
		})
	}
}

func TestNewTextManager(t *testing.T) {
	tests := []struct {
		name string
		file string
	}{
		{
			name: "OK",
			file: "VERSION",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			manager := NewTextManager(tc.file)
			assert.NotNil(t, manager)
		})
	}
}

func TestTextManagerRead(t *testing.T) {
	tests := []struct {
		name           string
		file           string
		expectedError  string
		expectedSemVer SemVer
	}{
		{
			name:          "NoFile",
			file:          "test/null",
			expectedError: "no such file or directory",
		},
		{
			name:          "EmptyFile",
			file:          "test/VERSION-empty",
			expectedError: "empty version file",
		},
		{
			name:          "InvalidVersion",
			file:          "test/VERSION-invalid",
			expectedError: "invalid semantic version",
		},
		{
			name:           "Success",
			file:           "test/VERSION",
			expectedSemVer: SemVer{Major: 0, Minor: 1, Patch: 0},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			manager := &textManager{
				file: tc.file,
			}

			semver, err := manager.Read()

			if tc.expectedError != "" {
				assert.Contains(t, err.Error(), tc.expectedError)
				assert.Empty(t, semver)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, semver, tc.expectedSemVer)
			}
		})
	}
}

func TestTextManagerUpdate(t *testing.T) {
	tests := []struct {
		name          string
		file          string
		version       string
		expectedError string
	}{
		{
			name:          "BadFile",
			file:          "/test/null",
			version:       "0.2.0",
			expectedError: "no such file or directory",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			manager := &textManager{
				file: tc.file,
			}

			err := manager.Update(tc.version)

			if tc.expectedError != "" {
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewJSONManager(t *testing.T) {
	tests := []struct {
		name string
		file string
	}{
		{
			name: "OK",
			file: "VERSION",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			manager := NewJSONManager(tc.file)
			assert.NotNil(t, manager)
		})
	}
}

func TestJSONManagerRead(t *testing.T) {
	tests := []struct {
		name           string
		file           string
		expectedError  string
		expectedSemVer SemVer
	}{
		{
			name:          "NoFile",
			file:          "test/null",
			expectedError: "no such file or directory",
		},
		{
			name:          "EmptyFile",
			file:          "test/package-empty.json",
			expectedError: "unexpected end of JSON input",
		},
		{
			name:          "NoVersionKey",
			file:          "test/package-no-version.json",
			expectedError: "no version key",
		},
		{
			name:          "BadVersionKey",
			file:          "test/package-bad-version.json",
			expectedError: "bad version key",
		},
		{
			name:          "InvalidVersionKey",
			file:          "test/package-invalid-version.json",
			expectedError: "invalid semantic version",
		},
		{
			name:           "Success",
			file:           "test/package.json",
			expectedSemVer: SemVer{Major: 0, Minor: 1, Patch: 0},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			manager := &jsonManager{
				file: tc.file,
			}

			semver, err := manager.Read()

			if tc.expectedError != "" {
				assert.Contains(t, err.Error(), tc.expectedError)
				assert.Empty(t, semver)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, semver, tc.expectedSemVer)
			}
		})
	}
}

func TestJSONManagerUpdate(t *testing.T) {
	tests := []struct {
		name          string
		file          string
		version       string
		expectedError string
	}{
		{
			name:          "BadFile",
			file:          "/test/null",
			version:       "0.2.0",
			expectedError: "no such file or directory",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			manager := &jsonManager{
				file: tc.file,
			}

			err := manager.Update(tc.version)

			if tc.expectedError != "" {
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
