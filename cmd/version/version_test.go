package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestString(t *testing.T) {
	tests := []struct {
		expectedString string
	}{
		{
			expectedString: "version:   revision:   branch:   goVersion:   buildTool:   buildTime: ",
		},
	}

	for _, tc := range tests {
		assert.Equal(t, tc.expectedString, String())
	}
}
