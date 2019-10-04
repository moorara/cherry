package action

import (
	"context"
	"errors"
	"testing"

	"github.com/moorara/cherry/internal/spec"
	"github.com/moorara/cherry/internal/step"
	"github.com/moorara/cherry/pkg/cui"
	"github.com/stretchr/testify/assert"
)

func TestNewBuild(t *testing.T) {
	tests := []struct {
		name    string
		ui      cui.CUI
		workDir string
		s       *spec.Spec
	}{
		{
			name:    "OK",
			ui:      &mockCUI{},
			workDir: "",
			s: &spec.Spec{
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
	}{
		{
			name: "Step1Fails",
			action: &build{
				ui:   &mockCUI{},
				spec: &spec.Spec{},
				step1: &step.GoList{
					Mock: &mockStep{
						RunOutError: errors.New("error on run: step1"),
					},
				},
			},
			expectedError: errors.New("error on run: step1"),
		},
		{
			name: "Step2Fails",
			action: &build{
				ui:   &mockCUI{},
				spec: &spec.Spec{},
				step1: &step.GoList{
					Mock: &mockStep{},
				},
				step2: &step.SemVerRead{
					Mock: &mockStep{
						RunOutError: errors.New("error on run: step2"),
					},
				},
			},
			expectedError: errors.New("error on run: step2"),
		},
		{
			name: "Step3Fails",
			action: &build{
				ui:   &mockCUI{},
				spec: &spec.Spec{},
				step1: &step.GoList{
					Mock: &mockStep{},
				},
				step2: &step.SemVerRead{
					Mock: &mockStep{},
				},
				step3: &step.GitGetHEAD{
					Mock: &mockStep{
						RunOutError: errors.New("error on run: step3"),
					},
				},
			},
			expectedError: errors.New("error on run: step3"),
		},
		{
			name: "Step4Fails",
			action: &build{
				ui:   &mockCUI{},
				spec: &spec.Spec{},
				step1: &step.GoList{
					Mock: &mockStep{},
				},
				step2: &step.SemVerRead{
					Mock: &mockStep{},
				},
				step3: &step.GitGetHEAD{
					Mock: &mockStep{},
				},
				step4: &step.GitGetBranch{
					Mock: &mockStep{
						RunOutError: errors.New("error on run: step4"),
					},
				},
			},
			expectedError: errors.New("error on run: step4"),
		},
		{
			name: "Step5Fails",
			action: &build{
				ui:   &mockCUI{},
				spec: &spec.Spec{},
				step1: &step.GoList{
					Mock: &mockStep{},
				},
				step2: &step.SemVerRead{
					Mock: &mockStep{},
				},
				step3: &step.GitGetHEAD{
					Mock: &mockStep{},
				},
				step4: &step.GitGetBranch{
					Mock: &mockStep{},
				},
				step5: &step.GoVersion{
					Mock: &mockStep{
						RunOutError: errors.New("error on run: step5"),
					},
				},
			},
			expectedError: errors.New("error on run: step5"),
		},
		{
			name: "Step6Fails",
			action: &build{
				ui:   &mockCUI{},
				spec: &spec.Spec{},
				step1: &step.GoList{
					Mock: &mockStep{},
				},
				step2: &step.SemVerRead{
					Mock: &mockStep{},
				},
				step3: &step.GitGetHEAD{
					Mock: &mockStep{},
				},
				step4: &step.GitGetBranch{
					Mock: &mockStep{},
				},
				step5: &step.GoVersion{
					Mock: &mockStep{},
				},
				step6: &step.GoBuild{
					Mock: &mockStep{
						DryOutError: errors.New("error on dry: step6"),
					},
				},
			},
			expectedError: errors.New("error on dry: step6"),
		},
		{
			name: "Success",
			action: &build{
				ui:   &mockCUI{},
				spec: &spec.Spec{},
				step1: &step.GoList{
					Mock: &mockStep{},
				},
				step2: &step.SemVerRead{
					Mock: &mockStep{},
				},
				step3: &step.GitGetHEAD{
					Mock: &mockStep{},
				},
				step4: &step.GitGetBranch{
					Mock: &mockStep{},
				},
				step5: &step.GoVersion{
					Mock: &mockStep{},
				},
				step6: &step.GoBuild{
					Mock: &mockStep{},
				},
			},
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

func TestBuildRun(t *testing.T) {
	tests := []struct {
		name          string
		action        Action
		ctx           context.Context
		expectedError error
	}{
		{
			name: "Step1Fails",
			action: &build{
				ui:   &mockCUI{},
				spec: &spec.Spec{},
				step1: &step.GoList{
					Mock: &mockStep{
						RunOutError: errors.New("error on run: step1"),
					},
				},
			},
			expectedError: errors.New("error on run: step1"),
		},
		{
			name: "Step2Fails",
			action: &build{
				ui:   &mockCUI{},
				spec: &spec.Spec{},
				step1: &step.GoList{
					Mock: &mockStep{},
				},
				step2: &step.SemVerRead{
					Mock: &mockStep{
						RunOutError: errors.New("error on run: step2"),
					},
				},
			},
			expectedError: errors.New("error on run: step2"),
		},
		{
			name: "Step3Fails",
			action: &build{
				ui:   &mockCUI{},
				spec: &spec.Spec{},
				step1: &step.GoList{
					Mock: &mockStep{},
				},
				step2: &step.SemVerRead{
					Mock: &mockStep{},
				},
				step3: &step.GitGetHEAD{
					Mock: &mockStep{
						RunOutError: errors.New("error on run: step3"),
					},
				},
			},
			expectedError: errors.New("error on run: step3"),
		},
		{
			name: "Step4Fails",
			action: &build{
				ui:   &mockCUI{},
				spec: &spec.Spec{},
				step1: &step.GoList{
					Mock: &mockStep{},
				},
				step2: &step.SemVerRead{
					Mock: &mockStep{},
				},
				step3: &step.GitGetHEAD{
					Mock: &mockStep{},
				},
				step4: &step.GitGetBranch{
					Mock: &mockStep{
						RunOutError: errors.New("error on run: step4"),
					},
				},
			},
			expectedError: errors.New("error on run: step4"),
		},
		{
			name: "Step5Fails",
			action: &build{
				ui:   &mockCUI{},
				spec: &spec.Spec{},
				step1: &step.GoList{
					Mock: &mockStep{},
				},
				step2: &step.SemVerRead{
					Mock: &mockStep{},
				},
				step3: &step.GitGetHEAD{
					Mock: &mockStep{},
				},
				step4: &step.GitGetBranch{
					Mock: &mockStep{},
				},
				step5: &step.GoVersion{
					Mock: &mockStep{
						RunOutError: errors.New("error on run: step5"),
					},
				},
			},
			expectedError: errors.New("error on run: step5"),
		},
		{
			name: "Step6Fails",
			action: &build{
				ui:   &mockCUI{},
				spec: &spec.Spec{},
				step1: &step.GoList{
					Mock: &mockStep{},
				},
				step2: &step.SemVerRead{
					Mock: &mockStep{},
				},
				step3: &step.GitGetHEAD{
					Mock: &mockStep{},
				},
				step4: &step.GitGetBranch{
					Mock: &mockStep{},
				},
				step5: &step.GoVersion{
					Mock: &mockStep{},
				},
				step6: &step.GoBuild{
					Mock: &mockStep{
						RunOutError: errors.New("error on run: step6"),
					},
				},
			},
			expectedError: errors.New("error on run: step6"),
		},
		{
			name: "Success",
			action: &build{
				ui:   &mockCUI{},
				spec: &spec.Spec{},
				step1: &step.GoList{
					Mock: &mockStep{},
				},
				step2: &step.SemVerRead{
					Mock: &mockStep{},
				},
				step3: &step.GitGetHEAD{
					Mock: &mockStep{},
				},
				step4: &step.GitGetBranch{
					Mock: &mockStep{},
				},
				step5: &step.GoVersion{
					Mock: &mockStep{},
				},
				step6: &step.GoBuild{
					Mock: &mockStep{},
				},
			},
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

func TestBuildRevert(t *testing.T) {
	tests := []struct {
		name          string
		action        Action
		ctx           context.Context
		expectedError error
	}{
		{
			name: "Step6Fails",
			action: &build{
				ui: &mockCUI{},
				step6: &step.GoBuild{
					Mock: &mockStep{
						RevertOutError: errors.New("error on revert: step6"),
					},
				},
			},
			expectedError: errors.New("error on revert: step6"),
		},
		{
			name: "Step5Fails",
			action: &build{
				ui: &mockCUI{},
				step6: &step.GoBuild{
					Mock: &mockStep{},
				},
				step5: &step.GoVersion{
					Mock: &mockStep{
						RevertOutError: errors.New("error on revert: step5"),
					},
				},
			},
			expectedError: errors.New("error on revert: step5"),
		},
		{
			name: "Step4Fails",
			action: &build{
				ui: &mockCUI{},
				step6: &step.GoBuild{
					Mock: &mockStep{},
				},
				step5: &step.GoVersion{
					Mock: &mockStep{},
				},
				step4: &step.GitGetBranch{
					Mock: &mockStep{
						RevertOutError: errors.New("error on revert: step4"),
					},
				},
			},
			expectedError: errors.New("error on revert: step4"),
		},
		{
			name: "Step3Fails",
			action: &build{
				ui: &mockCUI{},
				step6: &step.GoBuild{
					Mock: &mockStep{},
				},
				step5: &step.GoVersion{
					Mock: &mockStep{},
				},
				step4: &step.GitGetBranch{
					Mock: &mockStep{},
				},
				step3: &step.GitGetHEAD{
					Mock: &mockStep{
						RevertOutError: errors.New("error on revert: step3"),
					},
				},
			},
			expectedError: errors.New("error on revert: step3"),
		},
		{
			name: "Step2Fails",
			action: &build{
				ui: &mockCUI{},
				step6: &step.GoBuild{
					Mock: &mockStep{},
				},
				step5: &step.GoVersion{
					Mock: &mockStep{},
				},
				step4: &step.GitGetBranch{
					Mock: &mockStep{},
				},
				step3: &step.GitGetHEAD{
					Mock: &mockStep{},
				},
				step2: &step.SemVerRead{
					Mock: &mockStep{
						RevertOutError: errors.New("error on revert: step2"),
					},
				},
			},
			expectedError: errors.New("error on revert: step2"),
		},
		{
			name: "Step1Fails",
			action: &build{
				ui: &mockCUI{},
				step6: &step.GoBuild{
					Mock: &mockStep{},
				},
				step5: &step.GoVersion{
					Mock: &mockStep{},
				},
				step4: &step.GitGetBranch{
					Mock: &mockStep{},
				},
				step3: &step.GitGetHEAD{
					Mock: &mockStep{},
				},
				step2: &step.SemVerRead{
					Mock: &mockStep{},
				},
				step1: &step.GoList{
					Mock: &mockStep{
						RevertOutError: errors.New("error on revert: step1"),
					},
				},
			},
			expectedError: errors.New("error on revert: step1"),
		},
		{
			name: "Success",
			action: &build{
				ui: &mockCUI{},
				step6: &step.GoBuild{
					Mock: &mockStep{},
				},
				step5: &step.GoVersion{
					Mock: &mockStep{},
				},
				step4: &step.GitGetBranch{
					Mock: &mockStep{},
				},
				step3: &step.GitGetHEAD{
					Mock: &mockStep{},
				},
				step2: &step.SemVerRead{
					Mock: &mockStep{},
				},
				step1: &step.GoList{
					Mock: &mockStep{},
				},
			},
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
