package spec

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRead(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		expectedError string
		expectedSpec  *Spec
	}{
		{
			name:          "NoFile",
			path:          "test/null",
			expectedError: "no such file or directory",
			expectedSpec:  nil,
		},
		{
			name:          "EmptyYAML",
			path:          "test/empty.yaml",
			expectedError: "EOF",
			expectedSpec:  nil,
		},
		{
			name:          "InvalidYAML",
			path:          "test/error.yaml",
			expectedError: "cannot unmarshal",
			expectedSpec:  nil,
		},
		{
			"Success",
			"test/cherry.yaml",
			"",
			&Spec{
				Version: "1",
				Builds: []Build{
					Build{
						Language: "go",
						Version:  []string{"1.10", "1.11.2"},
						OS:       []string{"linux", "darwin", "windows"},
						Arch:     []string{"386", "amd64"},
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			spec, err := Read(tc.path)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Contains(t, err.Error(), tc.expectedError)
			}

			assert.Equal(t, tc.expectedSpec, spec)
		})
	}
}
