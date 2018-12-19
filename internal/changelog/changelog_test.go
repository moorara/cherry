package changelog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilename(t *testing.T) {
	tests := []struct {
		name             string
		workDir          string
		expectedFilename string
	}{

		{
			name:             "Success",
			workDir:          ".",
			expectedFilename: changelogFilename,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			changelog := New(tc.workDir)
			assert.NotNil(t, changelog)

			filename := changelog.Filename()
			assert.Equal(t, tc.expectedFilename, filename)
		})
	}
}
