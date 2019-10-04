package action

import (
	"context"
	"testing"

	"github.com/moorara/cherry/internal/spec"
	"github.com/moorara/cherry/pkg/cui"
	"github.com/stretchr/testify/assert"
)

func TestNewBuild(t *testing.T) {
	tests := []struct {
		name    string
		ui      cui.CUI
		workDir string
		s       spec.Spec
	}{
		{
			name:    "OK",
			ui:      &mockCUI{},
			workDir: "",
			s: spec.Spec{
				ToolName:    "cherry",
				ToolVersion: "test",
				Build: spec.Build{
					CrossCompile:   true,
					MainFile:       "main.go",
					BinaryFile:     "bin/app",
					VersionPackage: "./cmd/version",
					Platforms:      []string{"linux-amd64", "darwin-amd64"},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			action := NewBuild(tc.ui, tc.workDir, tc.s)
			assert.NotNil(t, action)
		})
	}
}

func TestBuildDry(t *testing.T) {
	tests := []struct {
		name          string
		action        Action
		ctx           context.Context
		expectedError error
	}{}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.action.Dry(tc.ctx)
			assert.Equal(t, tc.expectedError, err)
		})
	}
}

func TestBuildRun(t *testing.T) {
	tests := []struct {
		name          string
		action        Action
		ctx           context.Context
		expectedError error
	}{}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.action.Run(tc.ctx)
			assert.Equal(t, tc.expectedError, err)
		})
	}
}

func TestBuildRevert(t *testing.T) {
	tests := []struct {
		name          string
		action        Action
		ctx           context.Context
		expectedError error
	}{}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.action.Revert(tc.ctx)
			assert.Equal(t, tc.expectedError, err)
		})
	}
}
