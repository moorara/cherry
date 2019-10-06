package action

import (
	"context"
	"errors"
	"testing"

	"github.com/moorara/cherry/internal/step"
	"github.com/moorara/cherry/pkg/cui"
	"github.com/stretchr/testify/assert"
)

func TestNewUpdate(t *testing.T) {
	tests := []struct {
		name        string
		ui          cui.CUI
		githubToken string
	}{
		{
			name:        "OK",
			ui:          &mockCUI{},
			githubToken: "github-token",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			action := NewUpdate(tc.ui, tc.githubToken)
			assert.NotNil(t, action)
		})
	}
}

func TestUpdateDry(t *testing.T) {
	tests := []struct {
		name          string
		action        Action
		ctx           context.Context
		expectedError error
	}{
		{
			name: "Step1Fails",
			action: &update{
				ui: &mockCUI{},
				step1: &step.GitHubGetLatestRelease{
					Mock: &mockStep{
						RunOutError: errors.New("error on run: step1"),
					},
				},
			},
			ctx:           context.Background(),
			expectedError: errors.New("error on run: step1"),
		},
		{
			name: "Step2Fails",
			action: &update{
				ui: &mockCUI{},
				step1: &step.GitHubGetLatestRelease{
					Mock: &mockStep{},
				},
				step2: &step.GitHubDownloadAsset{
					Mock: &mockStep{
						DryOutError: errors.New("error on dry: step2"),
					},
				},
			},
			ctx:           context.Background(),
			expectedError: errors.New("error on dry: step2"),
		},
		{
			name: "Success",
			action: &update{
				ui: &mockCUI{},
				step1: &step.GitHubGetLatestRelease{
					Mock: &mockStep{},
				},
				step2: &step.GitHubDownloadAsset{
					Mock: &mockStep{},
				},
			},
			ctx:           context.Background(),
			expectedError: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.action.Dry(tc.ctx)
			assert.Equal(t, tc.expectedError, err)
		})
	}
}

func TestUpdateRun(t *testing.T) {
	tests := []struct {
		name          string
		action        Action
		ctx           context.Context
		expectedError error
	}{
		{
			name: "Step1Fails",
			action: &update{
				ui: &mockCUI{},
				step1: &step.GitHubGetLatestRelease{
					Mock: &mockStep{
						RunOutError: errors.New("error on run: step1"),
					},
				},
			},
			ctx:           context.Background(),
			expectedError: errors.New("error on run: step1"),
		},
		{
			name: "Step2Fails",
			action: &update{
				ui: &mockCUI{},
				step1: &step.GitHubGetLatestRelease{
					Mock: &mockStep{},
				},
				step2: &step.GitHubDownloadAsset{
					Mock: &mockStep{
						RunOutError: errors.New("error on run: step2"),
					},
				},
			},
			ctx:           context.Background(),
			expectedError: errors.New("error on run: step2"),
		},
		{
			name: "Success",
			action: &update{
				ui: &mockCUI{},
				step1: &step.GitHubGetLatestRelease{
					Mock: &mockStep{},
				},
				step2: &step.GitHubDownloadAsset{
					Mock: &mockStep{},
				},
			},
			ctx:           context.Background(),
			expectedError: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.action.Run(tc.ctx)
			assert.Equal(t, tc.expectedError, err)
		})
	}
}

func TestUpdateRevert(t *testing.T) {
	tests := []struct {
		name          string
		action        Action
		ctx           context.Context
		expectedError error
	}{
		{
			name: "Step2Fails",
			action: &update{
				ui: &mockCUI{},
				step2: &step.GitHubDownloadAsset{
					Mock: &mockStep{
						RevertOutError: errors.New("error on revert: step2"),
					},
				},
			},
			ctx:           context.Background(),
			expectedError: errors.New("error on revert: step2"),
		},
		{
			name: "Step1Fails",
			action: &update{
				ui: &mockCUI{},
				step2: &step.GitHubDownloadAsset{
					Mock: &mockStep{},
				},
				step1: &step.GitHubGetLatestRelease{
					Mock: &mockStep{
						RevertOutError: errors.New("error on revert: step1"),
					},
				},
			},
			ctx:           context.Background(),
			expectedError: errors.New("error on revert: step1"),
		},
		{
			name: "Success",
			action: &update{
				ui: &mockCUI{},
				step1: &step.GitHubGetLatestRelease{
					Mock: &mockStep{},
				},
				step2: &step.GitHubDownloadAsset{
					Mock: &mockStep{},
				},
			},
			ctx:           context.Background(),
			expectedError: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.action.Revert(tc.ctx)
			assert.Equal(t, tc.expectedError, err)
		})
	}
}
