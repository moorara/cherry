package model

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildDefaults(t *testing.T) {
	tests := []struct {
		name          string
		build         Build
		expectedBuild Build
	}{
		{
			"Empty",
			Build{},
			Build{
				Language: "",
				Version:  []string{},
				OS:       []string{},
				Arch:     []string{},
			},
		},
		{
			"DefaultsRequired",
			Build{
				Language: "go",
			},
			Build{
				Language: "go",
				Version:  []string{"1.11.2"},
				OS:       []string{runtime.GOOS},
				Arch:     []string{runtime.GOARCH},
			},
		},
		{
			"DefaultsNotRequired",
			Build{
				Language: "go",
				Version:  []string{"1.10", "1.11.2"},
				OS:       []string{"linux", "darwin", "windows"},
				Arch:     []string{"386", "amd64"},
			},
			Build{
				Language: "go",
				Version:  []string{"1.10", "1.11.2"},
				OS:       []string{"linux", "darwin", "windows"},
				Arch:     []string{"386", "amd64"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedBuild, tc.build.Defaults())
		})
	}
}
