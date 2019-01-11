package formula

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPackages(t *testing.T) {
	tests := []struct {
		name             string
		formula          *formula
		ctx              context.Context
		expectedError    string
		expectedPackages []string
	}{
		{
			name: "Error",
			formula: &formula{
				WorkDir: os.TempDir(),
			},
			ctx:              context.Background(),
			expectedError:    "can't load package",
			expectedPackages: nil,
		},
		{
			name: "Success",
			formula: &formula{
				WorkDir: ".",
			},
			ctx:           context.Background(),
			expectedError: "",
			expectedPackages: []string{
				"github.com/moorara/cherry/internal/v1/formula",
				"",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			packages, err := tc.formula.getPackages(tc.ctx)

			if tc.expectedError != "" {
				assert.Contains(t, err.Error(), tc.expectedError)
				assert.Nil(t, packages)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedPackages, packages)
			}
		})
	}
}

func TestTestPackage(t *testing.T) {
	tests := []struct {
		name    string
		formula *formula
	}{}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.NotNil(t, tc)
		})
	}
}

func TestCover(t *testing.T) {
	tests := []struct {
		name    string
		formula *formula
	}{}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.NotNil(t, tc)
		})
	}
}
