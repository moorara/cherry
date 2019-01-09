package spec

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReleaseSetDefaults(t *testing.T) {
	tests := []struct {
		release         Release
		expectedRelease Release
	}{
		{
			Release{},
			Release{
				Build: defaultBuild,
			},
		},
		{
			Release{
				Build: true,
			},
			Release{
				Build: true,
			},
		},
	}

	for _, tc := range tests {
		tc.release.SetDefaults()
		assert.Equal(t, tc.expectedRelease, tc.release)
	}
}

func TestReleaseFlagSet(t *testing.T) {
	tests := []struct {
		release      Release
		expectedName string
	}{
		{
			release:      Release{},
			expectedName: "release",
		},
		{
			release: Release{
				Build: true,
			},
			expectedName: "release",
		},
	}

	for _, tc := range tests {
		fs := tc.release.FlagSet()
		assert.Equal(t, tc.expectedName, fs.Name())
	}
}
