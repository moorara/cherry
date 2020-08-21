package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestString(t *testing.T) {
	tests := []struct {
		Version        string
		Commit         string
		Branch         string
		GoVersion      string
		BuildTool      string
		BuildTime      string
		expectedString string
	}{
		{
			Version:   "1.0.0",
			Commit:    "aaaaaaa",
			Branch:    "master",
			GoVersion: "go1.13",
			BuildTool: "tester",
			BuildTime: "2019-09-25T22:00:00",
			expectedString: `
	version:    1.0.0
	commit:     aaaaaaa
	branch:     master
	goVersion:  go1.13
	buildTool:  tester
	buildTime:  2019-09-25T22:00:00
`,
		},
	}

	for _, tc := range tests {
		Version = tc.Version
		Commit = tc.Commit
		Branch = tc.Branch
		GoVersion = tc.GoVersion
		BuildTool = tc.BuildTool
		BuildTime = tc.BuildTime

		assert.Contains(t, tc.expectedString, String())
	}
}
