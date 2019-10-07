package action

import (
	"context"
	"errors"
	"testing"

	"github.com/moorara/cherry/internal/spec"
	"github.com/moorara/cherry/internal/step"
	"github.com/moorara/cherry/pkg/cui"
	"github.com/moorara/cherry/pkg/semver"
	"github.com/stretchr/testify/assert"
)

func TestContextWithReleaseParams(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		segment semver.Segment
		comment string
	}{
		{
			name:    "Patch",
			ctx:     context.Background(),
			segment: semver.Patch,
			comment: "patch release",
		},
		{
			name:    "Minor",
			ctx:     context.Background(),
			segment: semver.Minor,
			comment: "minor release",
		},
		{
			name:    "Major",
			ctx:     context.Background(),
			segment: semver.Major,
			comment: "major release",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := ContextWithReleaseParams(tc.ctx, tc.segment, tc.comment)

			segment, ok := ctx.Value(segmentKey).(semver.Segment)
			assert.True(t, ok)
			assert.Equal(t, tc.segment, segment)

			comment, ok := ctx.Value(commentKey).(string)
			assert.True(t, ok)
			assert.Equal(t, tc.comment, comment)
		})
	}
}

func TestContext(t *testing.T) {
	tests := []struct {
		name            string
		ctx             context.Context
		expectedSegment semver.Segment
		expectedComment string
	}{
		{
			name:            "Default",
			ctx:             context.Background(),
			expectedSegment: semver.Patch,
			expectedComment: "",
		},
		{
			name:            "Patch",
			ctx:             context.WithValue(context.WithValue(context.Background(), segmentKey, semver.Patch), commentKey, "patch release"),
			expectedSegment: semver.Patch,
			expectedComment: "patch release",
		},
		{
			name:            "Minor",
			ctx:             context.WithValue(context.WithValue(context.Background(), segmentKey, semver.Minor), commentKey, "minor release"),
			expectedSegment: semver.Minor,
			expectedComment: "minor release",
		},
		{
			name:            "Major",
			ctx:             context.WithValue(context.WithValue(context.Background(), segmentKey, semver.Major), commentKey, "major release"),
			expectedSegment: semver.Major,
			expectedComment: "major release",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			segment, comment := ReleaseParamsFromContext(tc.ctx)
			assert.Equal(t, tc.expectedSegment, segment)
			assert.Equal(t, tc.expectedComment, comment)
		})
	}
}

func TestNewRelease(t *testing.T) {
	tests := []struct {
		name        string
		ui          cui.CUI
		workDir     string
		githubToken string
		s           spec.Spec
	}{
		{
			name:        "OK",
			ui:          &mockCUI{},
			workDir:     ".",
			githubToken: "github-token",
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
				Release: spec.Release{
					Model: "master",
					Build: true,
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			action := NewRelease(tc.ui, tc.workDir, tc.githubToken, tc.s)
			assert.NotNil(t, action)
		})
	}
}

func TestReleaseDry(t *testing.T) {
	ctx := ContextWithSpec(
		ContextWithReleaseParams(
			context.Background(),
			semver.Patch,
			"comment",
		),
		spec.Spec{
			ToolName:    "cherry",
			ToolVersion: "test",
			Build: spec.Build{
				Platforms: []string{"linux-amd64", "darwin-amd64"},
			},
			Release: spec.Release{
				Build: true,
			},
		},
	)

	step1OK := &step.GitGetRepo{Mock: &mockStep{}}

	step2OK := &step.GitGetBranch{Mock: &mockStep{}}
	step2OK.Result.Name = "master"

	step3OK := &step.GitStatus{Mock: &mockStep{}}
	step3OK.Result.IsClean = true

	step4OK := &step.GitPull{Mock: &mockStep{}}

	step5OK := &step.SemVerRead{Mock: &mockStep{}}
	step5OK.Result.Filename = "VERSION"
	step5OK.Result.Version = semver.SemVer{Major: 0, Minor: 2, Patch: 0}

	step6OK := &step.SemVerUpdate{Mock: &mockStep{}}
	step6OK.Result.Filename = "VERSION"

	step7OK := &step.GitHubCreateRelease{Mock: &mockStep{}}
	step8OK := &step.ChangelogGenerate{Mock: &mockStep{}}

	step9OK := &step.GitAdd{Mock: &mockStep{}}
	step10OK := &step.GitCommit{Mock: &mockStep{}}
	step11OK := &step.GitTag{Mock: &mockStep{}}

	step12OK := &step.GoList{Mock: &mockStep{}}
	step12OK.Result.PackagePath = "github.com/username/repo/cmd/version"

	step13OK := &step.GitGetHEAD{Mock: &mockStep{}}
	step13OK.Result.SHA = "aaaaaaa"
	step13OK.Result.ShortSHA = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

	step14OK := &step.GoVersion{Mock: &mockStep{}}
	step14OK.Result.Version = "go1.13"

	step15OK := &step.GoBuild{Mock: &mockStep{}}
	step15OK.Result.Binaries = []string{"/tmp/cherry-1234/bin/app"}

	step16OK := &step.GitHubUploadAssets{Mock: &mockStep{}}

	step17OK := &step.GitHubBranchProtection{Mock: &mockStep{}}
	step18OK := &step.GitHubBranchProtection{Mock: &mockStep{}}
	step19OK := &step.GitPush{Mock: &mockStep{}}
	step20OK := &step.GitPushTag{Mock: &mockStep{}}

	step21OK := &step.SemVerUpdate{Mock: &mockStep{}}
	step21OK.Result.Filename = "VERSION"

	step22OK := &step.GitAdd{Mock: &mockStep{}}
	step23OK := &step.GitCommit{Mock: &mockStep{}}
	step24OK := &step.GitPush{Mock: &mockStep{}}
	step25OK := &step.GitHubEditRelease{Mock: &mockStep{}}

	tests := []struct {
		name          string
		action        Action
		ctx           context.Context
		expectedError error
	}{
		{
			name: "Step1Fails",
			action: &release{
				ui: &mockCUI{},
				step1: &step.GitGetRepo{
					Mock: &mockStep{
						RunOutError: errors.New("error on run: step1"),
					},
				},
			},
			ctx:           ctx,
			expectedError: errors.New("error on run: step1"),
		},
		{
			name: "Step2Fails",
			action: &release{
				ui:    &mockCUI{},
				step1: step1OK,
				step2: &step.GitGetBranch{
					Mock: &mockStep{
						RunOutError: errors.New("error on run: step2"),
					},
				},
			},
			ctx:           ctx,
			expectedError: errors.New("error on run: step2"),
		},
		{
			name: "BranchNotMaster",
			action: &release{
				ui:    &mockCUI{},
				step1: step1OK,
				step2: &step.GitGetBranch{
					Mock: &mockStep{},
				},
			},
			ctx:           ctx,
			expectedError: errors.New("release has to be done from master branch"),
		},
		{
			name: "Step3Fails",
			action: &release{
				ui:    &mockCUI{},
				step1: step1OK,
				step2: step2OK,
				step3: &step.GitStatus{
					Mock: &mockStep{
						RunOutError: errors.New("error on run: step3"),
					},
				},
			},
			ctx:           ctx,
			expectedError: errors.New("error on run: step3"),
		},
		{
			name: "BranchNotClean",
			action: &release{
				ui:    &mockCUI{},
				step1: step1OK,
				step2: step2OK,
				step3: &step.GitStatus{
					Mock: &mockStep{},
				},
			},
			ctx:           ctx,
			expectedError: errors.New("working directory is not clean and has uncommitted changes"),
		},
		{
			name: "Step4Fails",
			action: &release{
				ui:    &mockCUI{},
				step1: step1OK,
				step2: step2OK,
				step3: step3OK,
				step4: &step.GitPull{
					Mock: &mockStep{
						DryOutError: errors.New("error on dry: step4"),
					},
				},
			},
			ctx:           ctx,
			expectedError: errors.New("error on dry: step4"),
		},
		{
			name: "Step5Fails",
			action: &release{
				ui:    &mockCUI{},
				step1: step1OK,
				step2: step2OK,
				step3: step3OK,
				step4: step4OK,
				step5: &step.SemVerRead{
					Mock: &mockStep{
						RunOutError: errors.New("error on run: step5"),
					},
				},
			},
			ctx:           ctx,
			expectedError: errors.New("error on run: step5"),
		},
		{
			name: "Step6Fails",
			action: &release{
				ui:    &mockCUI{},
				step1: step1OK,
				step2: step2OK,
				step3: step3OK,
				step4: step4OK,
				step5: step5OK,
				step6: &step.SemVerUpdate{
					Mock: &mockStep{
						DryOutError: errors.New("error on dry: step6"),
					},
				},
			},
			ctx:           ctx,
			expectedError: errors.New("error on dry: step6"),
		},
		{
			name: "Step7Fails",
			action: &release{
				ui:    &mockCUI{},
				step1: step1OK,
				step2: step2OK,
				step3: step3OK,
				step4: step4OK,
				step5: step5OK,
				step6: step6OK,
				step7: &step.GitHubCreateRelease{
					Mock: &mockStep{
						DryOutError: errors.New("error on dry: step7"),
					},
				},
			},
			ctx:           ctx,
			expectedError: errors.New("error on dry: step7"),
		},
		{
			name: "Step8Fails",
			action: &release{
				ui:    &mockCUI{},
				step1: step1OK,
				step2: step2OK,
				step3: step3OK,
				step4: step4OK,
				step5: step5OK,
				step6: step6OK,
				step7: step7OK,
				step8: &step.ChangelogGenerate{
					Mock: &mockStep{
						DryOutError: errors.New("error on dry: step8"),
					},
				},
			},
			ctx:           ctx,
			expectedError: errors.New("error on dry: step8"),
		},
		{
			name: "Step9Fails",
			action: &release{
				ui:    &mockCUI{},
				step1: step1OK,
				step2: step2OK,
				step3: step3OK,
				step4: step4OK,
				step5: step5OK,
				step6: step6OK,
				step7: step7OK,
				step8: step8OK,
				step9: &step.GitAdd{
					Mock: &mockStep{
						DryOutError: errors.New("error on dry: step9"),
					},
				},
			},
			ctx:           ctx,
			expectedError: errors.New("error on dry: step9"),
		},
		{
			name: "Step10Fails",
			action: &release{
				ui:    &mockCUI{},
				step1: step1OK,
				step2: step2OK,
				step3: step3OK,
				step4: step4OK,
				step5: step5OK,
				step6: step6OK,
				step7: step7OK,
				step8: step8OK,
				step9: step9OK,
				step10: &step.GitCommit{
					Mock: &mockStep{
						DryOutError: errors.New("error on dry: step10"),
					},
				},
			},
			ctx:           ctx,
			expectedError: errors.New("error on dry: step10"),
		},
		{
			name: "Step11Fails",
			action: &release{
				ui:     &mockCUI{},
				step1:  step1OK,
				step2:  step2OK,
				step3:  step3OK,
				step4:  step4OK,
				step5:  step5OK,
				step6:  step6OK,
				step7:  step7OK,
				step8:  step8OK,
				step9:  step9OK,
				step10: step10OK,
				step11: &step.GitTag{
					Mock: &mockStep{
						DryOutError: errors.New("error on dry: step11"),
					},
				},
			},
			ctx:           ctx,
			expectedError: errors.New("error on dry: step11"),
		},
		{
			name: "Step12Fails",
			action: &release{
				ui:     &mockCUI{},
				step1:  step1OK,
				step2:  step2OK,
				step3:  step3OK,
				step4:  step4OK,
				step5:  step5OK,
				step6:  step6OK,
				step7:  step7OK,
				step8:  step8OK,
				step9:  step9OK,
				step10: step10OK,
				step11: step11OK,
				step12: &step.GoList{
					Mock: &mockStep{
						RunOutError: errors.New("error on run: step12"),
					},
				},
			},
			ctx:           ctx,
			expectedError: errors.New("error on run: step12"),
		},
		{
			name: "Step13Fails",
			action: &release{
				ui:     &mockCUI{},
				step1:  step1OK,
				step2:  step2OK,
				step3:  step3OK,
				step4:  step4OK,
				step5:  step5OK,
				step6:  step6OK,
				step7:  step7OK,
				step8:  step8OK,
				step9:  step9OK,
				step10: step10OK,
				step11: step11OK,
				step12: step12OK,
				step13: &step.GitGetHEAD{
					Mock: &mockStep{
						RunOutError: errors.New("error on run: step13"),
					},
				},
			},
			ctx:           ctx,
			expectedError: errors.New("error on run: step13"),
		},
		{
			name: "Step14Fails",
			action: &release{
				ui:     &mockCUI{},
				step1:  step1OK,
				step2:  step2OK,
				step3:  step3OK,
				step4:  step4OK,
				step5:  step5OK,
				step6:  step6OK,
				step7:  step7OK,
				step8:  step8OK,
				step9:  step9OK,
				step10: step10OK,
				step11: step11OK,
				step12: step12OK,
				step13: step13OK,
				step14: &step.GoVersion{
					Mock: &mockStep{
						RunOutError: errors.New("error on run: step14"),
					},
				},
			},
			ctx:           ctx,
			expectedError: errors.New("error on run: step14"),
		},
		{
			name: "Step15Fails",
			action: &release{
				ui:     &mockCUI{},
				step1:  step1OK,
				step2:  step2OK,
				step3:  step3OK,
				step4:  step4OK,
				step5:  step5OK,
				step6:  step6OK,
				step7:  step7OK,
				step8:  step8OK,
				step9:  step9OK,
				step10: step10OK,
				step11: step11OK,
				step12: step12OK,
				step13: step13OK,
				step14: step14OK,
				step15: &step.GoBuild{
					Mock: &mockStep{
						DryOutError: errors.New("error on dry: step15"),
					},
				},
			},
			ctx:           ctx,
			expectedError: errors.New("error on dry: step15"),
		},
		{
			name: "Step16Fails",
			action: &release{
				ui:     &mockCUI{},
				step1:  step1OK,
				step2:  step2OK,
				step3:  step3OK,
				step4:  step4OK,
				step5:  step5OK,
				step6:  step6OK,
				step7:  step7OK,
				step8:  step8OK,
				step9:  step9OK,
				step10: step10OK,
				step11: step11OK,
				step12: step12OK,
				step13: step13OK,
				step14: step14OK,
				step15: step15OK,
				step16: &step.GitHubUploadAssets{
					Mock: &mockStep{
						DryOutError: errors.New("error on dry: step16"),
					},
				},
			},
			ctx:           ctx,
			expectedError: errors.New("error on dry: step16"),
		},
		{
			name: "Step17Fails",
			action: &release{
				ui:     &mockCUI{},
				step1:  step1OK,
				step2:  step2OK,
				step3:  step3OK,
				step4:  step4OK,
				step5:  step5OK,
				step6:  step6OK,
				step7:  step7OK,
				step8:  step8OK,
				step9:  step9OK,
				step10: step10OK,
				step11: step11OK,
				step12: step12OK,
				step13: step13OK,
				step14: step14OK,
				step15: step15OK,
				step16: step16OK,
				step17: &step.GitHubBranchProtection{
					Mock: &mockStep{
						DryOutError: errors.New("error on dry: step17"),
					},
				},
			},
			ctx:           ctx,
			expectedError: errors.New("error on dry: step17"),
		},
		{
			name: "Step19Fails",
			action: &release{
				ui:     &mockCUI{},
				step1:  step1OK,
				step2:  step2OK,
				step3:  step3OK,
				step4:  step4OK,
				step5:  step5OK,
				step6:  step6OK,
				step7:  step7OK,
				step8:  step8OK,
				step9:  step9OK,
				step10: step10OK,
				step11: step11OK,
				step12: step12OK,
				step13: step13OK,
				step14: step14OK,
				step15: step15OK,
				step16: step16OK,
				step17: step17OK,
				step18: step18OK,
				step19: &step.GitPush{
					Mock: &mockStep{
						DryOutError: errors.New("error on dry: step19"),
					},
				},
			},
			ctx:           ctx,
			expectedError: errors.New("error on dry: step19"),
		},
		{
			name: "Step20Fails",
			action: &release{
				ui:     &mockCUI{},
				step1:  step1OK,
				step2:  step2OK,
				step3:  step3OK,
				step4:  step4OK,
				step5:  step5OK,
				step6:  step6OK,
				step7:  step7OK,
				step8:  step8OK,
				step9:  step9OK,
				step10: step10OK,
				step11: step11OK,
				step12: step12OK,
				step13: step13OK,
				step14: step14OK,
				step15: step15OK,
				step16: step16OK,
				step17: step17OK,
				step18: step18OK,
				step19: step19OK,
				step20: &step.GitPushTag{
					Mock: &mockStep{
						DryOutError: errors.New("error on dry: step20"),
					},
				},
			},
			ctx:           ctx,
			expectedError: errors.New("error on dry: step20"),
		},
		{
			name: "Step21Fails",
			action: &release{
				ui:     &mockCUI{},
				step1:  step1OK,
				step2:  step2OK,
				step3:  step3OK,
				step4:  step4OK,
				step5:  step5OK,
				step6:  step6OK,
				step7:  step7OK,
				step8:  step8OK,
				step9:  step9OK,
				step10: step10OK,
				step11: step11OK,
				step12: step12OK,
				step13: step13OK,
				step14: step14OK,
				step15: step15OK,
				step16: step16OK,
				step17: step17OK,
				step18: step18OK,
				step19: step19OK,
				step20: step20OK,
				step21: &step.SemVerUpdate{
					Mock: &mockStep{
						DryOutError: errors.New("error on dry: step21"),
					},
				},
			},
			ctx:           ctx,
			expectedError: errors.New("error on dry: step21"),
		},
		{
			name: "Step22Fails",
			action: &release{
				ui:     &mockCUI{},
				step1:  step1OK,
				step2:  step2OK,
				step3:  step3OK,
				step4:  step4OK,
				step5:  step5OK,
				step6:  step6OK,
				step7:  step7OK,
				step8:  step8OK,
				step9:  step9OK,
				step10: step10OK,
				step11: step11OK,
				step12: step12OK,
				step13: step13OK,
				step14: step14OK,
				step15: step15OK,
				step16: step16OK,
				step17: step17OK,
				step18: step18OK,
				step19: step19OK,
				step20: step20OK,
				step21: step21OK,
				step22: &step.GitAdd{
					Mock: &mockStep{
						DryOutError: errors.New("error on dry: step22"),
					},
				},
			},
			ctx:           ctx,
			expectedError: errors.New("error on dry: step22"),
		},
		{
			name: "Step23Fails",
			action: &release{
				ui:     &mockCUI{},
				step1:  step1OK,
				step2:  step2OK,
				step3:  step3OK,
				step4:  step4OK,
				step5:  step5OK,
				step6:  step6OK,
				step7:  step7OK,
				step8:  step8OK,
				step9:  step9OK,
				step10: step10OK,
				step11: step11OK,
				step12: step12OK,
				step13: step13OK,
				step14: step14OK,
				step15: step15OK,
				step16: step16OK,
				step17: step17OK,
				step18: step18OK,
				step19: step19OK,
				step20: step20OK,
				step21: step21OK,
				step22: step22OK,
				step23: &step.GitCommit{
					Mock: &mockStep{
						DryOutError: errors.New("error on dry: step23"),
					},
				},
			},
			ctx:           ctx,
			expectedError: errors.New("error on dry: step23"),
		},
		{
			name: "Step24Fails",
			action: &release{
				ui:     &mockCUI{},
				step1:  step1OK,
				step2:  step2OK,
				step3:  step3OK,
				step4:  step4OK,
				step5:  step5OK,
				step6:  step6OK,
				step7:  step7OK,
				step8:  step8OK,
				step9:  step9OK,
				step10: step10OK,
				step11: step11OK,
				step12: step12OK,
				step13: step13OK,
				step14: step14OK,
				step15: step15OK,
				step16: step16OK,
				step17: step17OK,
				step18: step18OK,
				step19: step19OK,
				step20: step20OK,
				step21: step21OK,
				step22: step22OK,
				step23: step23OK,
				step24: &step.GitPush{
					Mock: &mockStep{
						DryOutError: errors.New("error on dry: step24"),
					},
				},
			},
			ctx:           ctx,
			expectedError: errors.New("error on dry: step24"),
		},
		{
			name: "Step25Fails",
			action: &release{
				ui:     &mockCUI{},
				step1:  step1OK,
				step2:  step2OK,
				step3:  step3OK,
				step4:  step4OK,
				step5:  step5OK,
				step6:  step6OK,
				step7:  step7OK,
				step8:  step8OK,
				step9:  step9OK,
				step10: step10OK,
				step11: step11OK,
				step12: step12OK,
				step13: step13OK,
				step14: step14OK,
				step15: step15OK,
				step16: step16OK,
				step17: step17OK,
				step18: step18OK,
				step19: step19OK,
				step20: step20OK,
				step21: step21OK,
				step22: step22OK,
				step23: step23OK,
				step24: step24OK,
				step25: &step.GitHubEditRelease{
					Mock: &mockStep{
						DryOutError: errors.New("error on dry: step25"),
					},
				},
			},
			ctx:           ctx,
			expectedError: errors.New("error on dry: step25"),
		},
		{
			name: "Success",
			action: &release{
				ui:     &mockCUI{},
				step1:  step1OK,
				step2:  step2OK,
				step3:  step3OK,
				step4:  step4OK,
				step5:  step5OK,
				step6:  step6OK,
				step7:  step7OK,
				step8:  step8OK,
				step9:  step9OK,
				step10: step10OK,
				step11: step11OK,
				step12: step12OK,
				step13: step13OK,
				step14: step14OK,
				step15: step15OK,
				step16: step16OK,
				step17: step17OK,
				step18: step18OK,
				step19: step19OK,
				step20: step20OK,
				step21: step21OK,
				step22: step22OK,
				step23: step23OK,
				step24: step24OK,
				step25: step25OK,
			},
			ctx: ctx,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.action.Dry(tc.ctx)
			assert.Equal(t, tc.expectedError, err)
		})
	}
}

func TestReleaseRun(t *testing.T) {
	ctx := ContextWithSpec(
		ContextWithReleaseParams(
			context.Background(),
			semver.Patch,
			"comment",
		),
		spec.Spec{
			ToolName:    "cherry",
			ToolVersion: "test",
			Build: spec.Build{
				Platforms: []string{"linux-amd64", "darwin-amd64"},
			},
			Release: spec.Release{
				Build: true,
			},
		},
	)

	step1OK := &step.GitGetRepo{Mock: &mockStep{}}

	step2OK := &step.GitGetBranch{Mock: &mockStep{}}
	step2OK.Result.Name = "master"

	step3OK := &step.GitStatus{Mock: &mockStep{}}
	step3OK.Result.IsClean = true

	step4OK := &step.GitPull{Mock: &mockStep{}}

	step5OK := &step.SemVerRead{Mock: &mockStep{}}
	step5OK.Result.Filename = "VERSION"
	step5OK.Result.Version = semver.SemVer{Major: 0, Minor: 2, Patch: 0}

	step6OK := &step.SemVerUpdate{Mock: &mockStep{}}
	step6OK.Result.Filename = "VERSION"

	step7OK := &step.GitHubCreateRelease{Mock: &mockStep{}}
	step7OK.Result.Release = step.GitHubRelease{
		ID:         2,
		Name:       "0.2.0",
		TagName:    "v0.2.0",
		Target:     "master",
		Draft:      true,
		Prerelease: false,
	}

	step8OK := &step.ChangelogGenerate{Mock: &mockStep{}}
	step8OK.Result.Filename = "CHANGELOG.md"
	step8OK.Result.Changelog = "change log ..."

	step9OK := &step.GitAdd{Mock: &mockStep{}}
	step10OK := &step.GitCommit{Mock: &mockStep{}}
	step11OK := &step.GitTag{Mock: &mockStep{}}

	step12OK := &step.GoList{Mock: &mockStep{}}
	step12OK.Result.PackagePath = "github.com/username/repo/cmd/version"

	step13OK := &step.GitGetHEAD{Mock: &mockStep{}}
	step13OK.Result.SHA = "aaaaaaa"
	step13OK.Result.ShortSHA = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

	step14OK := &step.GoVersion{Mock: &mockStep{}}
	step14OK.Result.Version = "go1.13"

	step15OK := &step.GoBuild{Mock: &mockStep{}}
	step15OK.Result.Binaries = []string{"bin/app-linux-amd64", "bin/app-darwin-amd64"}

	step16OK := &step.GitHubUploadAssets{Mock: &mockStep{}}
	step16OK.Result.Assets = []step.GitHubAsset{
		{ID: 1, Name: "bin/app-linux-amd64"},
		{ID: 2, Name: "bin/app-darwin-amd64"},
	}

	step17OK := &step.GitHubBranchProtection{Mock: &mockStep{}}
	step18OK := &step.GitHubBranchProtection{Mock: &mockStep{}}
	step19OK := &step.GitPush{Mock: &mockStep{}}
	step20OK := &step.GitPushTag{Mock: &mockStep{}}

	step21OK := &step.SemVerUpdate{Mock: &mockStep{}}
	step21OK.Result.Filename = "VERSION"

	step22OK := &step.GitAdd{Mock: &mockStep{}}
	step23OK := &step.GitCommit{Mock: &mockStep{}}
	step24OK := &step.GitPush{Mock: &mockStep{}}

	step25OK := &step.GitHubEditRelease{Mock: &mockStep{}}
	step25OK.Result.Release = step.GitHubRelease{
		ID:         2,
		Name:       "0.2.0",
		TagName:    "v0.2.0",
		Target:     "master",
		Draft:      false,
		Prerelease: false,
		Body:       "text",
	}

	tests := []struct {
		name          string
		action        Action
		ctx           context.Context
		expectedError error
	}{
		{
			name: "Step1Fails",
			action: &release{
				ui: &mockCUI{},
				step1: &step.GitGetRepo{
					Mock: &mockStep{
						RunOutError: errors.New("error on run: step1"),
					},
				},
			},
			ctx:           ctx,
			expectedError: errors.New("error on run: step1"),
		},
		{
			name: "Step2Fails",
			action: &release{
				ui:    &mockCUI{},
				step1: step1OK,
				step2: &step.GitGetBranch{
					Mock: &mockStep{
						RunOutError: errors.New("error on run: step2"),
					},
				},
			},
			ctx:           ctx,
			expectedError: errors.New("error on run: step2"),
		},
		{
			name: "BranchNotMaster",
			action: &release{
				ui:    &mockCUI{},
				step1: step1OK,
				step2: &step.GitGetBranch{
					Mock: &mockStep{},
				},
			},
			ctx:           ctx,
			expectedError: errors.New("release has to be done from master branch"),
		},
		{
			name: "Step3Fails",
			action: &release{
				ui:    &mockCUI{},
				step1: step1OK,
				step2: step2OK,
				step3: &step.GitStatus{
					Mock: &mockStep{
						RunOutError: errors.New("error on run: step3"),
					},
				},
			},
			ctx:           ctx,
			expectedError: errors.New("error on run: step3"),
		},
		{
			name: "BranchNotClean",
			action: &release{
				ui:    &mockCUI{},
				step1: step1OK,
				step2: step2OK,
				step3: &step.GitStatus{
					Mock: &mockStep{},
				},
			},
			ctx:           ctx,
			expectedError: errors.New("working directory is not clean and has uncommitted changes"),
		},
		{
			name: "Step4Fails",
			action: &release{
				ui:    &mockCUI{},
				step1: step1OK,
				step2: step2OK,
				step3: step3OK,
				step4: &step.GitPull{
					Mock: &mockStep{
						RunOutError: errors.New("error on run: step4"),
					},
				},
			},
			ctx:           ctx,
			expectedError: errors.New("error on run: step4"),
		},
		{
			name: "Step5Fails",
			action: &release{
				ui:    &mockCUI{},
				step1: step1OK,
				step2: step2OK,
				step3: step3OK,
				step4: step4OK,
				step5: &step.SemVerRead{
					Mock: &mockStep{
						RunOutError: errors.New("error on run: step5"),
					},
				},
			},
			ctx:           ctx,
			expectedError: errors.New("error on run: step5"),
		},
		{
			name: "Step6Fails",
			action: &release{
				ui:    &mockCUI{},
				step1: step1OK,
				step2: step2OK,
				step3: step3OK,
				step4: step4OK,
				step5: step5OK,
				step6: &step.SemVerUpdate{
					Mock: &mockStep{
						RunOutError: errors.New("error on run: step6"),
					},
				},
			},
			ctx:           ctx,
			expectedError: errors.New("error on run: step6"),
		},
		{
			name: "Step7Fails",
			action: &release{
				ui:    &mockCUI{},
				step1: step1OK,
				step2: step2OK,
				step3: step3OK,
				step4: step4OK,
				step5: step5OK,
				step6: step6OK,
				step7: &step.GitHubCreateRelease{
					Mock: &mockStep{
						RunOutError: errors.New("error on run: step7"),
					},
				},
			},
			ctx:           ctx,
			expectedError: errors.New("error on run: step7"),
		},
		{
			name: "Step8Fails",
			action: &release{
				ui:    &mockCUI{},
				step1: step1OK,
				step2: step2OK,
				step3: step3OK,
				step4: step4OK,
				step5: step5OK,
				step6: step6OK,
				step7: step7OK,
				step8: &step.ChangelogGenerate{
					Mock: &mockStep{
						RunOutError: errors.New("error on run: step8"),
					},
				},
			},
			ctx:           ctx,
			expectedError: errors.New("error on run: step8"),
		},
		{
			name: "Step9Fails",
			action: &release{
				ui:    &mockCUI{},
				step1: step1OK,
				step2: step2OK,
				step3: step3OK,
				step4: step4OK,
				step5: step5OK,
				step6: step6OK,
				step7: step7OK,
				step8: step8OK,
				step9: &step.GitAdd{
					Mock: &mockStep{
						RunOutError: errors.New("error on run: step9"),
					},
				},
			},
			ctx:           ctx,
			expectedError: errors.New("error on run: step9"),
		},
		{
			name: "Step10Fails",
			action: &release{
				ui:    &mockCUI{},
				step1: step1OK,
				step2: step2OK,
				step3: step3OK,
				step4: step4OK,
				step5: step5OK,
				step6: step6OK,
				step7: step7OK,
				step8: step8OK,
				step9: step9OK,
				step10: &step.GitCommit{
					Mock: &mockStep{
						RunOutError: errors.New("error on run: step10"),
					},
				},
			},
			ctx:           ctx,
			expectedError: errors.New("error on run: step10"),
		},
		{
			name: "Step11Fails",
			action: &release{
				ui:     &mockCUI{},
				step1:  step1OK,
				step2:  step2OK,
				step3:  step3OK,
				step4:  step4OK,
				step5:  step5OK,
				step6:  step6OK,
				step7:  step7OK,
				step8:  step8OK,
				step9:  step9OK,
				step10: step10OK,
				step11: &step.GitTag{
					Mock: &mockStep{
						RunOutError: errors.New("error on run: step11"),
					},
				},
			},
			ctx:           ctx,
			expectedError: errors.New("error on run: step11"),
		},
		{
			name: "Step12Fails",
			action: &release{
				ui:     &mockCUI{},
				step1:  step1OK,
				step2:  step2OK,
				step3:  step3OK,
				step4:  step4OK,
				step5:  step5OK,
				step6:  step6OK,
				step7:  step7OK,
				step8:  step8OK,
				step9:  step9OK,
				step10: step10OK,
				step11: step11OK,
				step12: &step.GoList{
					Mock: &mockStep{
						RunOutError: errors.New("error on run: step12"),
					},
				},
			},
			ctx:           ctx,
			expectedError: errors.New("error on run: step12"),
		},
		{
			name: "Step13Fails",
			action: &release{
				ui:     &mockCUI{},
				step1:  step1OK,
				step2:  step2OK,
				step3:  step3OK,
				step4:  step4OK,
				step5:  step5OK,
				step6:  step6OK,
				step7:  step7OK,
				step8:  step8OK,
				step9:  step9OK,
				step10: step10OK,
				step11: step11OK,
				step12: step12OK,
				step13: &step.GitGetHEAD{
					Mock: &mockStep{
						RunOutError: errors.New("error on run: step13"),
					},
				},
			},
			ctx:           ctx,
			expectedError: errors.New("error on run: step13"),
		},
		{
			name: "Step14Fails",
			action: &release{
				ui:     &mockCUI{},
				step1:  step1OK,
				step2:  step2OK,
				step3:  step3OK,
				step4:  step4OK,
				step5:  step5OK,
				step6:  step6OK,
				step7:  step7OK,
				step8:  step8OK,
				step9:  step9OK,
				step10: step10OK,
				step11: step11OK,
				step12: step12OK,
				step13: step13OK,
				step14: &step.GoVersion{
					Mock: &mockStep{
						RunOutError: errors.New("error on run: step14"),
					},
				},
			},
			ctx:           ctx,
			expectedError: errors.New("error on run: step14"),
		},
		{
			name: "Step15Fails",
			action: &release{
				ui:     &mockCUI{},
				step1:  step1OK,
				step2:  step2OK,
				step3:  step3OK,
				step4:  step4OK,
				step5:  step5OK,
				step6:  step6OK,
				step7:  step7OK,
				step8:  step8OK,
				step9:  step9OK,
				step10: step10OK,
				step11: step11OK,
				step12: step12OK,
				step13: step13OK,
				step14: step14OK,
				step15: &step.GoBuild{
					Mock: &mockStep{
						RunOutError: errors.New("error on run: step15"),
					},
				},
			},
			ctx:           ctx,
			expectedError: errors.New("error on run: step15"),
		},
		{
			name: "Step16Fails",
			action: &release{
				ui:     &mockCUI{},
				step1:  step1OK,
				step2:  step2OK,
				step3:  step3OK,
				step4:  step4OK,
				step5:  step5OK,
				step6:  step6OK,
				step7:  step7OK,
				step8:  step8OK,
				step9:  step9OK,
				step10: step10OK,
				step11: step11OK,
				step12: step12OK,
				step13: step13OK,
				step14: step14OK,
				step15: step15OK,
				step16: &step.GitHubUploadAssets{
					Mock: &mockStep{
						RunOutError: errors.New("error on run: step16"),
					},
				},
			},
			ctx:           ctx,
			expectedError: errors.New("error on run: step16"),
		},
		{
			name: "Step17Fails",
			action: &release{
				ui:     &mockCUI{},
				step1:  step1OK,
				step2:  step2OK,
				step3:  step3OK,
				step4:  step4OK,
				step5:  step5OK,
				step6:  step6OK,
				step7:  step7OK,
				step8:  step8OK,
				step9:  step9OK,
				step10: step10OK,
				step11: step11OK,
				step12: step12OK,
				step13: step13OK,
				step14: step14OK,
				step15: step15OK,
				step16: step16OK,
				step17: &step.GitHubBranchProtection{
					Mock: &mockStep{
						RunOutError: errors.New("error on run: step17"),
					},
				},
			},
			ctx:           ctx,
			expectedError: errors.New("error on run: step17"),
		},
		{
			name: "Step19Fails",
			action: &release{
				ui:     &mockCUI{},
				step1:  step1OK,
				step2:  step2OK,
				step3:  step3OK,
				step4:  step4OK,
				step5:  step5OK,
				step6:  step6OK,
				step7:  step7OK,
				step8:  step8OK,
				step9:  step9OK,
				step10: step10OK,
				step11: step11OK,
				step12: step12OK,
				step13: step13OK,
				step14: step14OK,
				step15: step15OK,
				step16: step16OK,
				step17: step17OK,
				step18: step18OK,
				step19: &step.GitPush{
					Mock: &mockStep{
						RunOutError: errors.New("error on run: step19"),
					},
				},
			},
			ctx:           ctx,
			expectedError: errors.New("error on run: step19"),
		},
		{
			name: "Step20Fails",
			action: &release{
				ui:     &mockCUI{},
				step1:  step1OK,
				step2:  step2OK,
				step3:  step3OK,
				step4:  step4OK,
				step5:  step5OK,
				step6:  step6OK,
				step7:  step7OK,
				step8:  step8OK,
				step9:  step9OK,
				step10: step10OK,
				step11: step11OK,
				step12: step12OK,
				step13: step13OK,
				step14: step14OK,
				step15: step15OK,
				step16: step16OK,
				step17: step17OK,
				step18: step18OK,
				step19: step19OK,
				step20: &step.GitPushTag{
					Mock: &mockStep{
						RunOutError: errors.New("error on run: step20"),
					},
				},
			},
			ctx:           ctx,
			expectedError: errors.New("error on run: step20"),
		},
		{
			name: "Step21Fails",
			action: &release{
				ui:     &mockCUI{},
				step1:  step1OK,
				step2:  step2OK,
				step3:  step3OK,
				step4:  step4OK,
				step5:  step5OK,
				step6:  step6OK,
				step7:  step7OK,
				step8:  step8OK,
				step9:  step9OK,
				step10: step10OK,
				step11: step11OK,
				step12: step12OK,
				step13: step13OK,
				step14: step14OK,
				step15: step15OK,
				step16: step16OK,
				step17: step17OK,
				step18: step18OK,
				step19: step19OK,
				step20: step20OK,
				step21: &step.SemVerUpdate{
					Mock: &mockStep{
						RunOutError: errors.New("error on run: step21"),
					},
				},
			},
			ctx:           ctx,
			expectedError: errors.New("error on run: step21"),
		},
		{
			name: "Step22Fails",
			action: &release{
				ui:     &mockCUI{},
				step1:  step1OK,
				step2:  step2OK,
				step3:  step3OK,
				step4:  step4OK,
				step5:  step5OK,
				step6:  step6OK,
				step7:  step7OK,
				step8:  step8OK,
				step9:  step9OK,
				step10: step10OK,
				step11: step11OK,
				step12: step12OK,
				step13: step13OK,
				step14: step14OK,
				step15: step15OK,
				step16: step16OK,
				step17: step17OK,
				step18: step18OK,
				step19: step19OK,
				step20: step20OK,
				step21: step21OK,
				step22: &step.GitAdd{
					Mock: &mockStep{
						RunOutError: errors.New("error on run: step22"),
					},
				},
			},
			ctx:           ctx,
			expectedError: errors.New("error on run: step22"),
		},
		{
			name: "Step23Fails",
			action: &release{
				ui:     &mockCUI{},
				step1:  step1OK,
				step2:  step2OK,
				step3:  step3OK,
				step4:  step4OK,
				step5:  step5OK,
				step6:  step6OK,
				step7:  step7OK,
				step8:  step8OK,
				step9:  step9OK,
				step10: step10OK,
				step11: step11OK,
				step12: step12OK,
				step13: step13OK,
				step14: step14OK,
				step15: step15OK,
				step16: step16OK,
				step17: step17OK,
				step18: step18OK,
				step19: step19OK,
				step20: step20OK,
				step21: step21OK,
				step22: step22OK,
				step23: &step.GitCommit{
					Mock: &mockStep{
						RunOutError: errors.New("error on run: step23"),
					},
				},
			},
			ctx:           ctx,
			expectedError: errors.New("error on run: step23"),
		},
		{
			name: "Step24Fails",
			action: &release{
				ui:     &mockCUI{},
				step1:  step1OK,
				step2:  step2OK,
				step3:  step3OK,
				step4:  step4OK,
				step5:  step5OK,
				step6:  step6OK,
				step7:  step7OK,
				step8:  step8OK,
				step9:  step9OK,
				step10: step10OK,
				step11: step11OK,
				step12: step12OK,
				step13: step13OK,
				step14: step14OK,
				step15: step15OK,
				step16: step16OK,
				step17: step17OK,
				step18: step18OK,
				step19: step19OK,
				step20: step20OK,
				step21: step21OK,
				step22: step22OK,
				step23: step23OK,
				step24: &step.GitPush{
					Mock: &mockStep{
						RunOutError: errors.New("error on run: step24"),
					},
				},
			},
			ctx:           ctx,
			expectedError: errors.New("error on run: step24"),
		},
		{
			name: "Step25Fails",
			action: &release{
				ui:     &mockCUI{},
				step1:  step1OK,
				step2:  step2OK,
				step3:  step3OK,
				step4:  step4OK,
				step5:  step5OK,
				step6:  step6OK,
				step7:  step7OK,
				step8:  step8OK,
				step9:  step9OK,
				step10: step10OK,
				step11: step11OK,
				step12: step12OK,
				step13: step13OK,
				step14: step14OK,
				step15: step15OK,
				step16: step16OK,
				step17: step17OK,
				step18: step18OK,
				step19: step19OK,
				step20: step20OK,
				step21: step21OK,
				step22: step22OK,
				step23: step23OK,
				step24: step24OK,
				step25: &step.GitHubEditRelease{
					Mock: &mockStep{
						RunOutError: errors.New("error on run: step25"),
					},
				},
			},
			ctx:           ctx,
			expectedError: errors.New("error on run: step25"),
		},
		{
			name: "Success",
			action: &release{
				ui:     &mockCUI{},
				step1:  step1OK,
				step2:  step2OK,
				step3:  step3OK,
				step4:  step4OK,
				step5:  step5OK,
				step6:  step6OK,
				step7:  step7OK,
				step8:  step8OK,
				step9:  step9OK,
				step10: step10OK,
				step11: step11OK,
				step12: step12OK,
				step13: step13OK,
				step14: step14OK,
				step15: step15OK,
				step16: step16OK,
				step17: step17OK,
				step18: step18OK,
				step19: step19OK,
				step20: step20OK,
				step21: step21OK,
				step22: step22OK,
				step23: step23OK,
				step24: step24OK,
				step25: step25OK,
			},
			ctx: ctx,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.action.Run(tc.ctx)
			assert.Equal(t, tc.expectedError, err)
		})
	}
}

func TestReleaseRevert(t *testing.T) {
	tests := []struct {
		name          string
		action        Action
		ctx           context.Context
		expectedError error
	}{
		{
			name: "Step25Fails",
			action: &release{
				ui: &mockCUI{},
				step25: &step.GitHubEditRelease{
					Mock: &mockStep{
						RevertOutError: errors.New("error on revert: step25"),
					},
				},
			},
			ctx:           context.Background(),
			expectedError: errors.New("error on revert: step25"),
		},
		{
			name: "Step24Fails",
			action: &release{
				ui: &mockCUI{},
				step25: &step.GitHubEditRelease{
					Mock: &mockStep{},
				},
				step24: &step.GitPush{
					Mock: &mockStep{
						RevertOutError: errors.New("error on revert: step24"),
					},
				},
			},
			ctx:           context.Background(),
			expectedError: errors.New("error on revert: step24"),
		},
		{
			name: "Step23Fails",
			action: &release{
				ui: &mockCUI{},
				step25: &step.GitHubEditRelease{
					Mock: &mockStep{},
				},
				step24: &step.GitPush{
					Mock: &mockStep{},
				},
				step23: &step.GitCommit{
					Mock: &mockStep{
						RevertOutError: errors.New("error on revert: step23"),
					},
				},
			},
			ctx:           context.Background(),
			expectedError: errors.New("error on revert: step23"),
		},
		{
			name: "Step22Fails",
			action: &release{
				ui: &mockCUI{},
				step25: &step.GitHubEditRelease{
					Mock: &mockStep{},
				},
				step24: &step.GitPush{
					Mock: &mockStep{},
				},
				step23: &step.GitCommit{
					Mock: &mockStep{},
				},
				step22: &step.GitAdd{
					Mock: &mockStep{
						RevertOutError: errors.New("error on revert: step22"),
					},
				},
			},
			ctx:           context.Background(),
			expectedError: errors.New("error on revert: step22"),
		},
		{
			name: "Step21Fails",
			action: &release{
				ui: &mockCUI{},
				step25: &step.GitHubEditRelease{
					Mock: &mockStep{},
				},
				step24: &step.GitPush{
					Mock: &mockStep{},
				},
				step23: &step.GitCommit{
					Mock: &mockStep{},
				},
				step22: &step.GitAdd{
					Mock: &mockStep{},
				},
				step21: &step.SemVerUpdate{
					Mock: &mockStep{
						RevertOutError: errors.New("error on revert: step21"),
					},
				},
			},
			ctx:           context.Background(),
			expectedError: errors.New("error on revert: step21"),
		},
		{
			name: "Step20Fails",
			action: &release{
				ui: &mockCUI{},
				step25: &step.GitHubEditRelease{
					Mock: &mockStep{},
				},
				step24: &step.GitPush{
					Mock: &mockStep{},
				},
				step23: &step.GitCommit{
					Mock: &mockStep{},
				},
				step22: &step.GitAdd{
					Mock: &mockStep{},
				},
				step21: &step.SemVerUpdate{
					Mock: &mockStep{},
				},
				step20: &step.GitPushTag{
					Mock: &mockStep{
						RevertOutError: errors.New("error on revert: step20"),
					},
				},
			},
			ctx:           context.Background(),
			expectedError: errors.New("error on revert: step20"),
		},
		{
			name: "Step19Fails",
			action: &release{
				ui: &mockCUI{},
				step25: &step.GitHubEditRelease{
					Mock: &mockStep{},
				},
				step24: &step.GitPush{
					Mock: &mockStep{},
				},
				step23: &step.GitCommit{
					Mock: &mockStep{},
				},
				step22: &step.GitAdd{
					Mock: &mockStep{},
				},
				step21: &step.SemVerUpdate{
					Mock: &mockStep{},
				},
				step20: &step.GitPushTag{
					Mock: &mockStep{},
				},
				step19: &step.GitPush{
					Mock: &mockStep{
						RevertOutError: errors.New("error on revert: step19"),
					},
				},
			},
			ctx:           context.Background(),
			expectedError: errors.New("error on revert: step19"),
		},
		{
			name: "Step18Fails",
			action: &release{
				ui: &mockCUI{},
				step25: &step.GitHubEditRelease{
					Mock: &mockStep{},
				},
				step24: &step.GitPush{
					Mock: &mockStep{},
				},
				step23: &step.GitCommit{
					Mock: &mockStep{},
				},
				step22: &step.GitAdd{
					Mock: &mockStep{},
				},
				step21: &step.SemVerUpdate{
					Mock: &mockStep{},
				},
				step20: &step.GitPushTag{
					Mock: &mockStep{},
				},
				step19: &step.GitPush{
					Mock: &mockStep{},
				},
				step18: &step.GitHubBranchProtection{
					Mock: &mockStep{
						RevertOutError: errors.New("error on revert: step18"),
					},
				},
			},
			ctx:           context.Background(),
			expectedError: errors.New("error on revert: step18"),
		},
		{
			name: "Step17Fails",
			action: &release{
				ui: &mockCUI{},
				step25: &step.GitHubEditRelease{
					Mock: &mockStep{},
				},
				step24: &step.GitPush{
					Mock: &mockStep{},
				},
				step23: &step.GitCommit{
					Mock: &mockStep{},
				},
				step22: &step.GitAdd{
					Mock: &mockStep{},
				},
				step21: &step.SemVerUpdate{
					Mock: &mockStep{},
				},
				step20: &step.GitPushTag{
					Mock: &mockStep{},
				},
				step19: &step.GitPush{
					Mock: &mockStep{},
				},
				step18: &step.GitHubBranchProtection{
					Mock: &mockStep{},
				},
				step17: &step.GitHubBranchProtection{
					Mock: &mockStep{
						RevertOutError: errors.New("error on revert: step17"),
					},
				},
			},
			ctx:           context.Background(),
			expectedError: errors.New("error on revert: step17"),
		},
		{
			name: "Step16Fails",
			action: &release{
				ui: &mockCUI{},
				step25: &step.GitHubEditRelease{
					Mock: &mockStep{},
				},
				step24: &step.GitPush{
					Mock: &mockStep{},
				},
				step23: &step.GitCommit{
					Mock: &mockStep{},
				},
				step22: &step.GitAdd{
					Mock: &mockStep{},
				},
				step21: &step.SemVerUpdate{
					Mock: &mockStep{},
				},
				step20: &step.GitPushTag{
					Mock: &mockStep{},
				},
				step19: &step.GitPush{
					Mock: &mockStep{},
				},
				step18: &step.GitHubBranchProtection{
					Mock: &mockStep{},
				},
				step17: &step.GitHubBranchProtection{
					Mock: &mockStep{},
				},
				step16: &step.GitHubUploadAssets{
					Mock: &mockStep{
						RevertOutError: errors.New("error on revert: step16"),
					},
				},
			},
			ctx:           context.Background(),
			expectedError: errors.New("error on revert: step16"),
		},
		{
			name: "Step15Fails",
			action: &release{
				ui: &mockCUI{},
				step25: &step.GitHubEditRelease{
					Mock: &mockStep{},
				},
				step24: &step.GitPush{
					Mock: &mockStep{},
				},
				step23: &step.GitCommit{
					Mock: &mockStep{},
				},
				step22: &step.GitAdd{
					Mock: &mockStep{},
				},
				step21: &step.SemVerUpdate{
					Mock: &mockStep{},
				},
				step20: &step.GitPushTag{
					Mock: &mockStep{},
				},
				step19: &step.GitPush{
					Mock: &mockStep{},
				},
				step18: &step.GitHubBranchProtection{
					Mock: &mockStep{},
				},
				step17: &step.GitHubBranchProtection{
					Mock: &mockStep{},
				},
				step16: &step.GitHubUploadAssets{
					Mock: &mockStep{},
				},
				step15: &step.GoBuild{
					Mock: &mockStep{
						RevertOutError: errors.New("error on revert: step15"),
					},
				},
			},
			ctx:           context.Background(),
			expectedError: errors.New("error on revert: step15"),
		},
		{
			name: "Step14Fails",
			action: &release{
				ui: &mockCUI{},
				step25: &step.GitHubEditRelease{
					Mock: &mockStep{},
				},
				step24: &step.GitPush{
					Mock: &mockStep{},
				},
				step23: &step.GitCommit{
					Mock: &mockStep{},
				},
				step22: &step.GitAdd{
					Mock: &mockStep{},
				},
				step21: &step.SemVerUpdate{
					Mock: &mockStep{},
				},
				step20: &step.GitPushTag{
					Mock: &mockStep{},
				},
				step19: &step.GitPush{
					Mock: &mockStep{},
				},
				step18: &step.GitHubBranchProtection{
					Mock: &mockStep{},
				},
				step17: &step.GitHubBranchProtection{
					Mock: &mockStep{},
				},
				step16: &step.GitHubUploadAssets{
					Mock: &mockStep{},
				},
				step15: &step.GoBuild{
					Mock: &mockStep{},
				},
				step14: &step.GoVersion{
					Mock: &mockStep{
						RevertOutError: errors.New("error on revert: step14"),
					},
				},
			},
			ctx:           context.Background(),
			expectedError: errors.New("error on revert: step14"),
		},
		{
			name: "Step13Fails",
			action: &release{
				ui: &mockCUI{},
				step25: &step.GitHubEditRelease{
					Mock: &mockStep{},
				},
				step24: &step.GitPush{
					Mock: &mockStep{},
				},
				step23: &step.GitCommit{
					Mock: &mockStep{},
				},
				step22: &step.GitAdd{
					Mock: &mockStep{},
				},
				step21: &step.SemVerUpdate{
					Mock: &mockStep{},
				},
				step20: &step.GitPushTag{
					Mock: &mockStep{},
				},
				step19: &step.GitPush{
					Mock: &mockStep{},
				},
				step18: &step.GitHubBranchProtection{
					Mock: &mockStep{},
				},
				step17: &step.GitHubBranchProtection{
					Mock: &mockStep{},
				},
				step16: &step.GitHubUploadAssets{
					Mock: &mockStep{},
				},
				step15: &step.GoBuild{
					Mock: &mockStep{},
				},
				step14: &step.GoVersion{
					Mock: &mockStep{},
				},
				step13: &step.GitGetHEAD{
					Mock: &mockStep{
						RevertOutError: errors.New("error on revert: step13"),
					},
				},
			},
			ctx:           context.Background(),
			expectedError: errors.New("error on revert: step13"),
		},
		{
			name: "Step12Fails",
			action: &release{
				ui: &mockCUI{},
				step25: &step.GitHubEditRelease{
					Mock: &mockStep{},
				},
				step24: &step.GitPush{
					Mock: &mockStep{},
				},
				step23: &step.GitCommit{
					Mock: &mockStep{},
				},
				step22: &step.GitAdd{
					Mock: &mockStep{},
				},
				step21: &step.SemVerUpdate{
					Mock: &mockStep{},
				},
				step20: &step.GitPushTag{
					Mock: &mockStep{},
				},
				step19: &step.GitPush{
					Mock: &mockStep{},
				},
				step18: &step.GitHubBranchProtection{
					Mock: &mockStep{},
				},
				step17: &step.GitHubBranchProtection{
					Mock: &mockStep{},
				},
				step16: &step.GitHubUploadAssets{
					Mock: &mockStep{},
				},
				step15: &step.GoBuild{
					Mock: &mockStep{},
				},
				step14: &step.GoVersion{
					Mock: &mockStep{},
				},
				step13: &step.GitGetHEAD{
					Mock: &mockStep{},
				},
				step12: &step.GoList{
					Mock: &mockStep{
						RevertOutError: errors.New("error on revert: step12"),
					},
				},
			},
			ctx:           context.Background(),
			expectedError: errors.New("error on revert: step12"),
		},
		{
			name: "Step11Fails",
			action: &release{
				ui: &mockCUI{},
				step25: &step.GitHubEditRelease{
					Mock: &mockStep{},
				},
				step24: &step.GitPush{
					Mock: &mockStep{},
				},
				step23: &step.GitCommit{
					Mock: &mockStep{},
				},
				step22: &step.GitAdd{
					Mock: &mockStep{},
				},
				step21: &step.SemVerUpdate{
					Mock: &mockStep{},
				},
				step20: &step.GitPushTag{
					Mock: &mockStep{},
				},
				step19: &step.GitPush{
					Mock: &mockStep{},
				},
				step18: &step.GitHubBranchProtection{
					Mock: &mockStep{},
				},
				step17: &step.GitHubBranchProtection{
					Mock: &mockStep{},
				},
				step16: &step.GitHubUploadAssets{
					Mock: &mockStep{},
				},
				step15: &step.GoBuild{
					Mock: &mockStep{},
				},
				step14: &step.GoVersion{
					Mock: &mockStep{},
				},
				step13: &step.GitGetHEAD{
					Mock: &mockStep{},
				},
				step12: &step.GoList{
					Mock: &mockStep{},
				},
				step11: &step.GitTag{
					Mock: &mockStep{
						RevertOutError: errors.New("error on revert: step11"),
					},
				},
			},
			ctx:           context.Background(),
			expectedError: errors.New("error on revert: step11"),
		},
		{
			name: "Step10Fails",
			action: &release{
				ui: &mockCUI{},
				step25: &step.GitHubEditRelease{
					Mock: &mockStep{},
				},
				step24: &step.GitPush{
					Mock: &mockStep{},
				},
				step23: &step.GitCommit{
					Mock: &mockStep{},
				},
				step22: &step.GitAdd{
					Mock: &mockStep{},
				},
				step21: &step.SemVerUpdate{
					Mock: &mockStep{},
				},
				step20: &step.GitPushTag{
					Mock: &mockStep{},
				},
				step19: &step.GitPush{
					Mock: &mockStep{},
				},
				step18: &step.GitHubBranchProtection{
					Mock: &mockStep{},
				},
				step17: &step.GitHubBranchProtection{
					Mock: &mockStep{},
				},
				step16: &step.GitHubUploadAssets{
					Mock: &mockStep{},
				},
				step15: &step.GoBuild{
					Mock: &mockStep{},
				},
				step14: &step.GoVersion{
					Mock: &mockStep{},
				},
				step13: &step.GitGetHEAD{
					Mock: &mockStep{},
				},
				step12: &step.GoList{
					Mock: &mockStep{},
				},
				step11: &step.GitTag{
					Mock: &mockStep{},
				},
				step10: &step.GitCommit{
					Mock: &mockStep{
						RevertOutError: errors.New("error on revert: step10"),
					},
				},
			},
			ctx:           context.Background(),
			expectedError: errors.New("error on revert: step10"),
		},
		{
			name: "Step9Fails",
			action: &release{
				ui: &mockCUI{},
				step25: &step.GitHubEditRelease{
					Mock: &mockStep{},
				},
				step24: &step.GitPush{
					Mock: &mockStep{},
				},
				step23: &step.GitCommit{
					Mock: &mockStep{},
				},
				step22: &step.GitAdd{
					Mock: &mockStep{},
				},
				step21: &step.SemVerUpdate{
					Mock: &mockStep{},
				},
				step20: &step.GitPushTag{
					Mock: &mockStep{},
				},
				step19: &step.GitPush{
					Mock: &mockStep{},
				},
				step18: &step.GitHubBranchProtection{
					Mock: &mockStep{},
				},
				step17: &step.GitHubBranchProtection{
					Mock: &mockStep{},
				},
				step16: &step.GitHubUploadAssets{
					Mock: &mockStep{},
				},
				step15: &step.GoBuild{
					Mock: &mockStep{},
				},
				step14: &step.GoVersion{
					Mock: &mockStep{},
				},
				step13: &step.GitGetHEAD{
					Mock: &mockStep{},
				},
				step12: &step.GoList{
					Mock: &mockStep{},
				},
				step11: &step.GitTag{
					Mock: &mockStep{},
				},
				step10: &step.GitCommit{
					Mock: &mockStep{},
				},
				step9: &step.GitAdd{
					Mock: &mockStep{
						RevertOutError: errors.New("error on revert: step9"),
					},
				},
			},
			ctx:           context.Background(),
			expectedError: errors.New("error on revert: step9"),
		},
		{
			name: "Step8Fails",
			action: &release{
				ui: &mockCUI{},
				step25: &step.GitHubEditRelease{
					Mock: &mockStep{},
				},
				step24: &step.GitPush{
					Mock: &mockStep{},
				},
				step23: &step.GitCommit{
					Mock: &mockStep{},
				},
				step22: &step.GitAdd{
					Mock: &mockStep{},
				},
				step21: &step.SemVerUpdate{
					Mock: &mockStep{},
				},
				step20: &step.GitPushTag{
					Mock: &mockStep{},
				},
				step19: &step.GitPush{
					Mock: &mockStep{},
				},
				step18: &step.GitHubBranchProtection{
					Mock: &mockStep{},
				},
				step17: &step.GitHubBranchProtection{
					Mock: &mockStep{},
				},
				step16: &step.GitHubUploadAssets{
					Mock: &mockStep{},
				},
				step15: &step.GoBuild{
					Mock: &mockStep{},
				},
				step14: &step.GoVersion{
					Mock: &mockStep{},
				},
				step13: &step.GitGetHEAD{
					Mock: &mockStep{},
				},
				step12: &step.GoList{
					Mock: &mockStep{},
				},
				step11: &step.GitTag{
					Mock: &mockStep{},
				},
				step10: &step.GitCommit{
					Mock: &mockStep{},
				},
				step9: &step.GitAdd{
					Mock: &mockStep{},
				},
				step8: &step.ChangelogGenerate{
					Mock: &mockStep{
						RevertOutError: errors.New("error on revert: step8"),
					},
				},
			},
			ctx:           context.Background(),
			expectedError: errors.New("error on revert: step8"),
		},
		{
			name: "Step7Fails",
			action: &release{
				ui: &mockCUI{},
				step25: &step.GitHubEditRelease{
					Mock: &mockStep{},
				},
				step24: &step.GitPush{
					Mock: &mockStep{},
				},
				step23: &step.GitCommit{
					Mock: &mockStep{},
				},
				step22: &step.GitAdd{
					Mock: &mockStep{},
				},
				step21: &step.SemVerUpdate{
					Mock: &mockStep{},
				},
				step20: &step.GitPushTag{
					Mock: &mockStep{},
				},
				step19: &step.GitPush{
					Mock: &mockStep{},
				},
				step18: &step.GitHubBranchProtection{
					Mock: &mockStep{},
				},
				step17: &step.GitHubBranchProtection{
					Mock: &mockStep{},
				},
				step16: &step.GitHubUploadAssets{
					Mock: &mockStep{},
				},
				step15: &step.GoBuild{
					Mock: &mockStep{},
				},
				step14: &step.GoVersion{
					Mock: &mockStep{},
				},
				step13: &step.GitGetHEAD{
					Mock: &mockStep{},
				},
				step12: &step.GoList{
					Mock: &mockStep{},
				},
				step11: &step.GitTag{
					Mock: &mockStep{},
				},
				step10: &step.GitCommit{
					Mock: &mockStep{},
				},
				step9: &step.GitAdd{
					Mock: &mockStep{},
				},
				step8: &step.ChangelogGenerate{
					Mock: &mockStep{},
				},
				step7: &step.GitHubCreateRelease{
					Mock: &mockStep{
						RevertOutError: errors.New("error on revert: step7"),
					},
				},
			},
			ctx:           context.Background(),
			expectedError: errors.New("error on revert: step7"),
		},
		{
			name: "Step6Fails",
			action: &release{
				ui: &mockCUI{},
				step25: &step.GitHubEditRelease{
					Mock: &mockStep{},
				},
				step24: &step.GitPush{
					Mock: &mockStep{},
				},
				step23: &step.GitCommit{
					Mock: &mockStep{},
				},
				step22: &step.GitAdd{
					Mock: &mockStep{},
				},
				step21: &step.SemVerUpdate{
					Mock: &mockStep{},
				},
				step20: &step.GitPushTag{
					Mock: &mockStep{},
				},
				step19: &step.GitPush{
					Mock: &mockStep{},
				},
				step18: &step.GitHubBranchProtection{
					Mock: &mockStep{},
				},
				step17: &step.GitHubBranchProtection{
					Mock: &mockStep{},
				},
				step16: &step.GitHubUploadAssets{
					Mock: &mockStep{},
				},
				step15: &step.GoBuild{
					Mock: &mockStep{},
				},
				step14: &step.GoVersion{
					Mock: &mockStep{},
				},
				step13: &step.GitGetHEAD{
					Mock: &mockStep{},
				},
				step12: &step.GoList{
					Mock: &mockStep{},
				},
				step11: &step.GitTag{
					Mock: &mockStep{},
				},
				step10: &step.GitCommit{
					Mock: &mockStep{},
				},
				step9: &step.GitAdd{
					Mock: &mockStep{},
				},
				step8: &step.ChangelogGenerate{
					Mock: &mockStep{},
				},
				step7: &step.GitHubCreateRelease{
					Mock: &mockStep{},
				},
				step6: &step.SemVerUpdate{
					Mock: &mockStep{
						RevertOutError: errors.New("error on revert: step6"),
					},
				},
			},
			ctx:           context.Background(),
			expectedError: errors.New("error on revert: step6"),
		},
		{
			name: "Step5Fails",
			action: &release{
				ui: &mockCUI{},
				step25: &step.GitHubEditRelease{
					Mock: &mockStep{},
				},
				step24: &step.GitPush{
					Mock: &mockStep{},
				},
				step23: &step.GitCommit{
					Mock: &mockStep{},
				},
				step22: &step.GitAdd{
					Mock: &mockStep{},
				},
				step21: &step.SemVerUpdate{
					Mock: &mockStep{},
				},
				step20: &step.GitPushTag{
					Mock: &mockStep{},
				},
				step19: &step.GitPush{
					Mock: &mockStep{},
				},
				step18: &step.GitHubBranchProtection{
					Mock: &mockStep{},
				},
				step17: &step.GitHubBranchProtection{
					Mock: &mockStep{},
				},
				step16: &step.GitHubUploadAssets{
					Mock: &mockStep{},
				},
				step15: &step.GoBuild{
					Mock: &mockStep{},
				},
				step14: &step.GoVersion{
					Mock: &mockStep{},
				},
				step13: &step.GitGetHEAD{
					Mock: &mockStep{},
				},
				step12: &step.GoList{
					Mock: &mockStep{},
				},
				step11: &step.GitTag{
					Mock: &mockStep{},
				},
				step10: &step.GitCommit{
					Mock: &mockStep{},
				},
				step9: &step.GitAdd{
					Mock: &mockStep{},
				},
				step8: &step.ChangelogGenerate{
					Mock: &mockStep{},
				},
				step7: &step.GitHubCreateRelease{
					Mock: &mockStep{},
				},
				step6: &step.SemVerUpdate{
					Mock: &mockStep{},
				},
				step5: &step.SemVerRead{
					Mock: &mockStep{
						RevertOutError: errors.New("error on revert: step5"),
					},
				},
			},
			ctx:           context.Background(),
			expectedError: errors.New("error on revert: step5"),
		},
		{
			name: "Step4Fails",
			action: &release{
				ui: &mockCUI{},
				step25: &step.GitHubEditRelease{
					Mock: &mockStep{},
				},
				step24: &step.GitPush{
					Mock: &mockStep{},
				},
				step23: &step.GitCommit{
					Mock: &mockStep{},
				},
				step22: &step.GitAdd{
					Mock: &mockStep{},
				},
				step21: &step.SemVerUpdate{
					Mock: &mockStep{},
				},
				step20: &step.GitPushTag{
					Mock: &mockStep{},
				},
				step19: &step.GitPush{
					Mock: &mockStep{},
				},
				step18: &step.GitHubBranchProtection{
					Mock: &mockStep{},
				},
				step17: &step.GitHubBranchProtection{
					Mock: &mockStep{},
				},
				step16: &step.GitHubUploadAssets{
					Mock: &mockStep{},
				},
				step15: &step.GoBuild{
					Mock: &mockStep{},
				},
				step14: &step.GoVersion{
					Mock: &mockStep{},
				},
				step13: &step.GitGetHEAD{
					Mock: &mockStep{},
				},
				step12: &step.GoList{
					Mock: &mockStep{},
				},
				step11: &step.GitTag{
					Mock: &mockStep{},
				},
				step10: &step.GitCommit{
					Mock: &mockStep{},
				},
				step9: &step.GitAdd{
					Mock: &mockStep{},
				},
				step8: &step.ChangelogGenerate{
					Mock: &mockStep{},
				},
				step7: &step.GitHubCreateRelease{
					Mock: &mockStep{},
				},
				step6: &step.SemVerUpdate{
					Mock: &mockStep{},
				},
				step5: &step.SemVerRead{
					Mock: &mockStep{},
				},
				step4: &step.GitPull{
					Mock: &mockStep{
						RevertOutError: errors.New("error on revert: step4"),
					},
				},
			},
			ctx:           context.Background(),
			expectedError: errors.New("error on revert: step4"),
		},
		{
			name: "Step3Fails",
			action: &release{
				ui: &mockCUI{},
				step25: &step.GitHubEditRelease{
					Mock: &mockStep{},
				},
				step24: &step.GitPush{
					Mock: &mockStep{},
				},
				step23: &step.GitCommit{
					Mock: &mockStep{},
				},
				step22: &step.GitAdd{
					Mock: &mockStep{},
				},
				step21: &step.SemVerUpdate{
					Mock: &mockStep{},
				},
				step20: &step.GitPushTag{
					Mock: &mockStep{},
				},
				step19: &step.GitPush{
					Mock: &mockStep{},
				},
				step18: &step.GitHubBranchProtection{
					Mock: &mockStep{},
				},
				step17: &step.GitHubBranchProtection{
					Mock: &mockStep{},
				},
				step16: &step.GitHubUploadAssets{
					Mock: &mockStep{},
				},
				step15: &step.GoBuild{
					Mock: &mockStep{},
				},
				step14: &step.GoVersion{
					Mock: &mockStep{},
				},
				step13: &step.GitGetHEAD{
					Mock: &mockStep{},
				},
				step12: &step.GoList{
					Mock: &mockStep{},
				},
				step11: &step.GitTag{
					Mock: &mockStep{},
				},
				step10: &step.GitCommit{
					Mock: &mockStep{},
				},
				step9: &step.GitAdd{
					Mock: &mockStep{},
				},
				step8: &step.ChangelogGenerate{
					Mock: &mockStep{},
				},
				step7: &step.GitHubCreateRelease{
					Mock: &mockStep{},
				},
				step6: &step.SemVerUpdate{
					Mock: &mockStep{},
				},
				step5: &step.SemVerRead{
					Mock: &mockStep{},
				},
				step4: &step.GitPull{
					Mock: &mockStep{},
				},
				step3: &step.GitStatus{
					Mock: &mockStep{
						RevertOutError: errors.New("error on revert: step3"),
					},
				},
			},
			ctx:           context.Background(),
			expectedError: errors.New("error on revert: step3"),
		},
		{
			name: "Step2Fails",
			action: &release{
				ui: &mockCUI{},
				step25: &step.GitHubEditRelease{
					Mock: &mockStep{},
				},
				step24: &step.GitPush{
					Mock: &mockStep{},
				},
				step23: &step.GitCommit{
					Mock: &mockStep{},
				},
				step22: &step.GitAdd{
					Mock: &mockStep{},
				},
				step21: &step.SemVerUpdate{
					Mock: &mockStep{},
				},
				step20: &step.GitPushTag{
					Mock: &mockStep{},
				},
				step19: &step.GitPush{
					Mock: &mockStep{},
				},
				step18: &step.GitHubBranchProtection{
					Mock: &mockStep{},
				},
				step17: &step.GitHubBranchProtection{
					Mock: &mockStep{},
				},
				step16: &step.GitHubUploadAssets{
					Mock: &mockStep{},
				},
				step15: &step.GoBuild{
					Mock: &mockStep{},
				},
				step14: &step.GoVersion{
					Mock: &mockStep{},
				},
				step13: &step.GitGetHEAD{
					Mock: &mockStep{},
				},
				step12: &step.GoList{
					Mock: &mockStep{},
				},
				step11: &step.GitTag{
					Mock: &mockStep{},
				},
				step10: &step.GitCommit{
					Mock: &mockStep{},
				},
				step9: &step.GitAdd{
					Mock: &mockStep{},
				},
				step8: &step.ChangelogGenerate{
					Mock: &mockStep{},
				},
				step7: &step.GitHubCreateRelease{
					Mock: &mockStep{},
				},
				step6: &step.SemVerUpdate{
					Mock: &mockStep{},
				},
				step5: &step.SemVerRead{
					Mock: &mockStep{},
				},
				step4: &step.GitPull{
					Mock: &mockStep{},
				},
				step3: &step.GitStatus{
					Mock: &mockStep{},
				},
				step2: &step.GitGetBranch{
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
			action: &release{
				ui: &mockCUI{},
				step25: &step.GitHubEditRelease{
					Mock: &mockStep{},
				},
				step24: &step.GitPush{
					Mock: &mockStep{},
				},
				step23: &step.GitCommit{
					Mock: &mockStep{},
				},
				step22: &step.GitAdd{
					Mock: &mockStep{},
				},
				step21: &step.SemVerUpdate{
					Mock: &mockStep{},
				},
				step20: &step.GitPushTag{
					Mock: &mockStep{},
				},
				step19: &step.GitPush{
					Mock: &mockStep{},
				},
				step18: &step.GitHubBranchProtection{
					Mock: &mockStep{},
				},
				step17: &step.GitHubBranchProtection{
					Mock: &mockStep{},
				},
				step16: &step.GitHubUploadAssets{
					Mock: &mockStep{},
				},
				step15: &step.GoBuild{
					Mock: &mockStep{},
				},
				step14: &step.GoVersion{
					Mock: &mockStep{},
				},
				step13: &step.GitGetHEAD{
					Mock: &mockStep{},
				},
				step12: &step.GoList{
					Mock: &mockStep{},
				},
				step11: &step.GitTag{
					Mock: &mockStep{},
				},
				step10: &step.GitCommit{
					Mock: &mockStep{},
				},
				step9: &step.GitAdd{
					Mock: &mockStep{},
				},
				step8: &step.ChangelogGenerate{
					Mock: &mockStep{},
				},
				step7: &step.GitHubCreateRelease{
					Mock: &mockStep{},
				},
				step6: &step.SemVerUpdate{
					Mock: &mockStep{},
				},
				step5: &step.SemVerRead{
					Mock: &mockStep{},
				},
				step4: &step.GitPull{
					Mock: &mockStep{},
				},
				step3: &step.GitStatus{
					Mock: &mockStep{},
				},
				step2: &step.GitGetBranch{
					Mock: &mockStep{},
				},
				step1: &step.GitGetRepo{
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
			action: &release{
				ui: &mockCUI{},
				step25: &step.GitHubEditRelease{
					Mock: &mockStep{},
				},
				step24: &step.GitPush{
					Mock: &mockStep{},
				},
				step23: &step.GitCommit{
					Mock: &mockStep{},
				},
				step22: &step.GitAdd{
					Mock: &mockStep{},
				},
				step21: &step.SemVerUpdate{
					Mock: &mockStep{},
				},
				step20: &step.GitPushTag{
					Mock: &mockStep{},
				},
				step19: &step.GitPush{
					Mock: &mockStep{},
				},
				step18: &step.GitHubBranchProtection{
					Mock: &mockStep{},
				},
				step17: &step.GitHubBranchProtection{
					Mock: &mockStep{},
				},
				step16: &step.GitHubUploadAssets{
					Mock: &mockStep{},
				},
				step15: &step.GoBuild{
					Mock: &mockStep{},
				},
				step14: &step.GoVersion{
					Mock: &mockStep{},
				},
				step13: &step.GitGetHEAD{
					Mock: &mockStep{},
				},
				step12: &step.GoList{
					Mock: &mockStep{},
				},
				step11: &step.GitTag{
					Mock: &mockStep{},
				},
				step10: &step.GitCommit{
					Mock: &mockStep{},
				},
				step9: &step.GitAdd{
					Mock: &mockStep{},
				},
				step8: &step.ChangelogGenerate{
					Mock: &mockStep{},
				},
				step7: &step.GitHubCreateRelease{
					Mock: &mockStep{},
				},
				step6: &step.SemVerUpdate{
					Mock: &mockStep{},
				},
				step5: &step.SemVerRead{
					Mock: &mockStep{},
				},
				step4: &step.GitPull{
					Mock: &mockStep{},
				},
				step3: &step.GitStatus{
					Mock: &mockStep{},
				},
				step2: &step.GitGetBranch{
					Mock: &mockStep{},
				},
				step1: &step.GitGetRepo{
					Mock: &mockStep{},
				},
			},
			ctx: context.Background(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.action.Revert(tc.ctx)
			assert.Equal(t, tc.expectedError, err)
		})
	}
}
