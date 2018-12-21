package spec

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTestSetDefaults(t *testing.T) {
	tests := []struct {
		test         Test
		expectedTest Test
	}{
		{
			Test{},
			Test{
				ReportPath: defaultReportPath,
			},
		},
		{
			Test{
				ReportPath: "report",
			},
			Test{
				ReportPath: "report",
			},
		},
	}

	for _, tc := range tests {
		tc.test.SetDefaults()
		assert.Equal(t, tc.expectedTest, tc.test)
	}
}

func TestTestFlagSet(t *testing.T) {
	tests := []struct {
		test         Test
		expectedName string
	}{
		{
			test:         Test{},
			expectedName: "test",
		},
		{
			test: Test{
				ReportPath: "coverage",
			},
			expectedName: "test",
		},
	}

	for _, tc := range tests {
		fs := tc.test.FlagSet()
		assert.Equal(t, tc.expectedName, fs.Name())
	}
}
