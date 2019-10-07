package action

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/moorara/cherry/internal/spec"
	"github.com/moorara/cherry/internal/step"
	"github.com/moorara/cherry/pkg/cui"
	"github.com/moorara/cherry/pkg/semver"
)

const (
	segmentKey = contextKey("ReleaseSegment")
	commentKey = contextKey("ReleaseComment")
)

// ContextWithReleaseParams returns a new context that has input parameters for Release action.
func ContextWithReleaseParams(ctx context.Context, segment semver.Segment, comment string) context.Context {
	ctx = context.WithValue(ctx, segmentKey, segment)
	ctx = context.WithValue(ctx, commentKey, comment)

	return ctx
}

// ReleaseParamsFromContext retrieves input parameters for Release action from a context.
// If a parameter is not found, a default value will be returned.
func ReleaseParamsFromContext(ctx context.Context) (segment semver.Segment, comment string) {
	var ok bool

	segment, ok = ctx.Value(segmentKey).(semver.Segment)
	if !ok {
		segment = semver.Patch
	}

	comment, ok = ctx.Value(commentKey).(string)
	if !ok {
		comment = ""
	}

	return segment, comment
}

// release is the action for release command.
type release struct {
	ui     cui.CUI
	step1  *step.GitGetRepo
	step2  *step.GitGetBranch
	step3  *step.GitStatus
	step4  *step.GitPull
	step5  *step.SemVerRead
	step6  *step.SemVerUpdate
	step7  *step.GitHubCreateRelease
	step8  *step.ChangelogGenerate
	step9  *step.GitAdd
	step10 *step.GitCommit
	step11 *step.GitTag
	step12 *step.GoList
	step13 *step.GitGetHEAD
	step14 *step.GoVersion
	step15 *step.GoBuild
	step16 *step.GitHubUploadAssets
	step17 *step.GitHubBranchProtection
	step18 *step.GitHubBranchProtection
	step19 *step.GitPush
	step20 *step.GitPushTag
	step21 *step.SemVerUpdate
	step22 *step.GitAdd
	step23 *step.GitCommit
	step24 *step.GitPush
	step25 *step.GitHubEditRelease
}

// NewRelease creates an instance of Release action.
func NewRelease(ui cui.CUI, workDir, githubToken string, s spec.Spec) Action {
	transport := &http.Transport{}
	client := &http.Client{
		Transport: transport,
	}

	return &release{
		ui: ui,
		step1: &step.GitGetRepo{
			WorkDir: workDir,
		},
		step2: &step.GitGetBranch{
			WorkDir: workDir,
		},
		step3: &step.GitStatus{
			WorkDir: workDir,
		},
		step4: &step.GitPull{
			WorkDir: workDir,
		},
		step5: &step.SemVerRead{
			WorkDir:  workDir,
			Filename: s.VersionFile,
		},
		step6: &step.SemVerUpdate{
			WorkDir:  workDir,
			Filename: s.VersionFile,
			Version:  "TBD",
		},
		step7: &step.GitHubCreateRelease{
			Client:  client,
			Token:   githubToken,
			BaseURL: step.GitHubAPIURL,
			Repo:    "TBD",
			ReleaseData: step.GitHubReleaseData{
				Name:       "TBD",
				TagName:    "TBD",
				Target:     "TBD",
				Draft:      true,
				Prerelease: false,
			},
		},
		step8: &step.ChangelogGenerate{
			WorkDir:     workDir,
			GitHubToken: githubToken,
			Repo:        "TBD",
			Tag:         "TBD",
		},
		step9: &step.GitAdd{
			WorkDir: workDir,
			Files:   nil, // TBD
		},
		step10: &step.GitCommit{
			WorkDir: workDir,
			Message: "TBD",
		},
		step11: &step.GitTag{
			WorkDir:    workDir,
			Tag:        "TBD",
			Annotation: "TBD",
		},
		step12: &step.GoList{
			WorkDir: workDir,
			Package: s.Build.VersionPackage,
		},
		step13: &step.GitGetHEAD{
			WorkDir: workDir,
		},
		step14: &step.GoVersion{
			WorkDir: workDir,
		},
		step15: &step.GoBuild{
			WorkDir:    workDir,
			LDFlags:    "TBD",
			MainFile:   s.Build.MainFile,
			BinaryFile: s.Build.BinaryFile,
			Platforms:  nil, // TBD
		},
		step16: &step.GitHubUploadAssets{
			Client:           client,
			Token:            githubToken,
			BaseURL:          step.GitHubAPIURL,
			Repo:             "TBD",
			ReleaseID:        0, // TBD
			ReleaseUploadURL: "TBD",
			AssetFiles:       nil, // TBD
		},
		step17: &step.GitHubBranchProtection{
			Client:  client,
			Token:   githubToken,
			BaseURL: step.GitHubAPIURL,
			Repo:    "TBD",
			Branch:  "TBD",
			Enabled: false,
		},
		step18: &step.GitHubBranchProtection{
			Client:  client,
			Token:   githubToken,
			BaseURL: step.GitHubAPIURL,
			Repo:    "TBD",
			Branch:  "TBD",
			Enabled: true,
		},
		step19: &step.GitPush{
			WorkDir: workDir,
		},
		step20: &step.GitPushTag{
			WorkDir: workDir,
			Tag:     "TBD",
		},
		step21: &step.SemVerUpdate{
			WorkDir:  workDir,
			Filename: s.VersionFile,
			Version:  "TBD",
		},
		step22: &step.GitAdd{
			WorkDir: workDir,
			Files:   nil, // TBD
		},
		step23: &step.GitCommit{
			WorkDir: workDir,
			Message: "TBD",
		},
		step24: &step.GitPush{
			WorkDir: workDir,
		},
		step25: &step.GitHubEditRelease{
			Client:    client,
			Token:     githubToken,
			BaseURL:   step.GitHubAPIURL,
			Repo:      "TBD",
			ReleaseID: 0, // TBD
			ReleaseData: step.GitHubReleaseData{
				Name:       "TBD",
				TagName:    "TBD",
				Target:     "TBD",
				Draft:      false,
				Prerelease: false,
				Body:       "TBD",
			},
		},
	}
}

func (r *release) getLDFlags(s spec.Spec) string {
	buildTool := s.ToolName
	if s.ToolVersion != "" {
		buildTool += "@" + s.ToolVersion
	}

	vPkg := r.step12.Result.PackagePath
	versionFlag := fmt.Sprintf("-X %s.Version=%s", vPkg, r.step6.Version)
	revisionFlag := fmt.Sprintf("-X %s.Revision=%s", vPkg, r.step13.Result.ShortSHA)
	branchFlag := fmt.Sprintf("-X %s.Branch=%s", vPkg, r.step2.Result.Name)
	goVersionFlag := fmt.Sprintf("-X %s.GoVersion=%s", vPkg, r.step14.Result.Version)
	buildToolFlag := fmt.Sprintf("-X %s.BuildTool=%s", vPkg, buildTool)
	buildTimeFlag := fmt.Sprintf("-X %s.BuildTime=%s", vPkg, time.Now().UTC().Format(time.RFC3339Nano))
	ldflags := fmt.Sprintf("%s %s %s %s %s %s", versionFlag, revisionFlag, branchFlag, goVersionFlag, buildToolFlag, buildTimeFlag)

	return ldflags
}

// Dry is a dry run of the action.
func (r *release) Dry(ctx context.Context) error {
	s := SpecFromContext(ctx)
	segment, _ := ReleaseParamsFromContext(ctx)

	// Get repo name
	if err := r.step1.Run(ctx); err != nil {
		return err
	}

	// Get branch name
	if err := r.step2.Run(ctx); err != nil {
		return err
	}

	if r.step2.Result.Name != "master" {
		return errors.New("release has to be done from master branch")
	}

	// Get git status
	if err := r.step3.Run(ctx); err != nil {
		return err
	}

	if !r.step3.Result.IsClean {
		return errors.New("working directory is not clean and has uncommitted changes")
	}

	// Dry -- Pulling master branch
	if err := r.step4.Dry(ctx); err != nil {
		return err
	}

	// Read the version
	if err := r.step5.Run(ctx); err != nil {
		return err
	}

	// Release the version
	curr, next := r.step5.Result.Version.Release(segment)
	next.Prerelease = []string{"0"}

	// Dry -- Update the version file with the current version
	r.step6.Version = curr.Version()
	if err := r.step6.Dry(ctx); err != nil {
		return err
	}

	// Dry -- Create a draft release
	r.step7.Repo = r.step1.Result.Repo
	if err := r.step7.Dry(ctx); err != nil {
		return err
	}

	// Dry -- Create/Update change log
	r.step8.Repo = r.step1.Result.Repo
	r.step8.Tag = curr.GitTag()
	if err := r.step8.Dry(ctx); err != nil {
		return err
	}

	// Dry -- Add unstaged to files to staging
	r.step9.Files = []string{r.step6.Result.Filename, r.step8.Result.Filename}
	if err := r.step9.Dry(ctx); err != nil {
		return err
	}

	// Dry -- Create a commit for current version
	r.step10.Message = fmt.Sprintf("Releasing %s", curr.Version())
	if err := r.step10.Dry(ctx); err != nil {
		return err
	}

	// Dry -- Create a tag for current version
	if err := r.step11.Dry(ctx); err != nil {
		return err
	}

	if s.Release.Build {
		// Find package version path
		if err := r.step12.Run(ctx); err != nil {
			return err
		}

		// Get commit SHA hashes
		if err := r.step13.Run(ctx); err != nil {
			return err
		}

		// Get Go version
		if err := r.step14.Run(ctx); err != nil {
			return err
		}

		// Dry -- Cross-compile and build artifacts
		r.step15.LDFlags = r.getLDFlags(s)
		if err := r.step15.Dry(ctx); err != nil {
			return err
		}

		// Dry -- Upload build artifacts to release
		r.step16.Repo = r.step1.Result.Repo
		r.step16.ReleaseID = r.step7.Result.Release.ID
		if err := r.step16.Dry(ctx); err != nil {
			return err
		}
	}

	// Dry -- Temporarily disable the master branch protection
	r.step17.Repo = r.step1.Result.Repo
	r.step17.Branch = r.step2.Result.Name
	if err := r.step17.Dry(ctx); err != nil {
		return err
	}

	// Dry -- Make sure we re-enable the master branch protection
	defer func() {
		r.step18.Repo = r.step1.Result.Repo
		r.step18.Branch = r.step2.Result.Name
		if err := r.step18.Dry(ctx); err != nil {
			// Skip
		}
	}()

	// Dry -- Push the commit for current release
	if err := r.step19.Dry(ctx); err != nil {
		return err
	}

	// Dry -- Push the tag for current release
	if err := r.step20.Dry(ctx); err != nil {
		return err
	}

	// Dry -- Update the version file with the next version
	r.step21.Version = next.Version()
	if err := r.step21.Dry(ctx); err != nil {
		return err
	}

	// Dry -- Add unstaged to files to staging
	r.step22.Files = []string{r.step21.Result.Filename}
	if err := r.step22.Dry(ctx); err != nil {
		return err
	}

	//  Dry -- Create a commit for next version
	r.step23.Message = fmt.Sprintf("Beginning %s [skip ci]", next.Version())
	if err := r.step23.Dry(ctx); err != nil {
		return err
	}

	// Dry -- Push the commit for next release
	if err := r.step24.Dry(ctx); err != nil {
		return err
	}

	// Dry -- Edit the draft release and make it ready
	r.step25.Repo = r.step1.Result.Repo
	r.step25.ReleaseID = r.step7.Result.Release.ID
	if err := r.step25.Dry(ctx); err != nil {
		return err
	}

	return nil
}

// Run executes the action.
func (r *release) Run(ctx context.Context) error {
	s := SpecFromContext(ctx)
	segment, comment := ReleaseParamsFromContext(ctx)

	// Get repo name
	if err := r.step1.Run(ctx); err != nil {
		return err
	}

	// Get branch name
	if err := r.step2.Run(ctx); err != nil {
		return err
	}

	if r.step2.Result.Name != "master" {
		return errors.New("release has to be done from master branch")
	}

	// Get git status
	if err := r.step3.Run(ctx); err != nil {
		return err
	}

	// This is to ensure that we do not commit any unwanted change while releasing
	if !r.step3.Result.IsClean {
		return errors.New("working directory is not clean and has uncommitted changes")
	}

	r.ui.Outputf("⬇️  Pulling master branch ...")

	// Pulling master branch
	if err := r.step4.Run(ctx); err != nil {
		return err
	}

	// Read the version
	if err := r.step5.Run(ctx); err != nil {
		return err
	}

	// Release the version
	curr, next := r.step5.Result.Version.Release(segment)
	next.Prerelease = []string{"0"}

	// Update the version file with the current version
	r.step6.Version = curr.Version()
	if err := r.step6.Run(ctx); err != nil {
		return err
	}

	r.ui.Outputf("⬆️  Creating draft release %s ...", curr.Version())

	// Create a draft release
	r.step7.Repo = r.step1.Result.Repo
	r.step7.ReleaseData.Name = curr.Version()
	r.step7.ReleaseData.TagName = curr.GitTag()
	r.step7.ReleaseData.Target = r.step2.Result.Name
	if err := r.step7.Run(ctx); err != nil {
		return err
	}

	r.ui.Outputf("➡️  Creating/Updating change log ...")

	// Create/Update change log
	r.step8.Repo = r.step1.Result.Repo
	r.step8.Tag = curr.GitTag()
	if err := r.step8.Run(ctx); err != nil {
		return err
	}

	// Add unstaged to files to staging
	r.step9.Files = []string{r.step6.Result.Filename, r.step8.Result.Filename}
	if err := r.step9.Run(ctx); err != nil {
		return err
	}

	// Create a commit for current version
	r.step10.Message = fmt.Sprintf("Releasing %s", curr.Version())
	if err := r.step10.Run(ctx); err != nil {
		return err
	}

	// Create a tag for current version
	r.step11.Tag = curr.GitTag()
	r.step11.Annotation = fmt.Sprintf("Version %s", curr.Version())
	if err := r.step11.Run(ctx); err != nil {
		return err
	}

	if s.Release.Build {
		r.ui.Outputf("➡️  Building artifacts ...")

		// Find package version path
		if err := r.step12.Run(ctx); err != nil {
			return err
		}

		// Get commit SHA hashes
		if err := r.step13.Run(ctx); err != nil {
			return err
		}

		// Get Go version
		if err := r.step14.Run(ctx); err != nil {
			return err
		}

		// Cross-compile and build artifacts
		r.step15.LDFlags = r.getLDFlags(s)
		r.step15.Platforms = s.Build.Platforms
		if err := r.step15.Run(ctx); err != nil {
			return err
		}

		r.ui.Outputf("➡️️  Uploading artifacts to release %s ...", r.step7.Result.Release.Name)

		// Upload build artifacts to release
		r.step16.Repo = r.step1.Result.Repo
		r.step16.ReleaseID = r.step7.Result.Release.ID
		r.step16.ReleaseUploadURL = r.step7.Result.Release.UploadURL
		r.step16.AssetFiles = r.step15.Result.Binaries
		if err := r.step16.Run(ctx); err != nil {
			return err
		}
	}

	r.ui.Warnf("🔓 Temporarily enabling push to master branch ...")

	// Temporarily disable the master branch protection
	r.step17.Repo = r.step1.Result.Repo
	r.step17.Branch = r.step2.Result.Name
	if err := r.step17.Run(ctx); err != nil {
		return err
	}

	// Make sure we re-enable the master branch protection
	defer func() {
		r.ui.Warnf("🔒 Re-disabling push to master branch ...")

		r.step18.Repo = r.step1.Result.Repo
		r.step18.Branch = r.step2.Result.Name
		if err := r.step18.Run(ctx); err != nil {
			r.ui.Errorf("Error: %s", err)
		}
	}()

	r.ui.Infof("⬆️  Pushing release commit %s ...", r.step7.Result.Release.Name)

	// Push the commit for current release
	if err := r.step19.Run(ctx); err != nil {
		return err
	}

	r.ui.Infof("⬆️  Pushing release tag %s ...", r.step7.Result.Release.Name)

	// Push the tag for current release
	r.step20.Tag = curr.GitTag()
	if err := r.step20.Run(ctx); err != nil {
		return err
	}

	// Update the version file with the next version
	r.step21.Version = next.Version()
	if err := r.step21.Run(ctx); err != nil {
		return err
	}

	// Add unstaged to files to staging
	r.step22.Files = []string{r.step21.Result.Filename}
	if err := r.step22.Run(ctx); err != nil {
		return err
	}

	//  Create a commit for next version
	r.step23.Message = fmt.Sprintf("Beginning %s [skip ci]", next.Version())
	if err := r.step23.Run(ctx); err != nil {
		return err
	}

	r.ui.Infof("⬆️  Pushing commit for next version %s ...", next.Version())

	// Push the commit for next release
	if err := r.step24.Run(ctx); err != nil {
		return err
	}

	r.ui.Infof("⬆️  Publishing release %s ...", r.step7.Result.Release.Name)

	// Edit the draft release and make it ready
	r.step25.Repo = r.step1.Result.Repo
	r.step25.ReleaseID = r.step7.Result.Release.ID
	r.step25.ReleaseData.Name = curr.Version()
	r.step25.ReleaseData.TagName = curr.GitTag()
	r.step25.ReleaseData.Target = r.step2.Result.Name
	r.step25.ReleaseData.Body = fmt.Sprintf("%s\n\n%s", comment, r.step8.Result.Changelog)
	if err := r.step25.Run(ctx); err != nil {
		return err
	}

	return nil
}

// Revert reverts back an executed action.
func (r *release) Revert(ctx context.Context) error {
	steps := []step.Step{
		r.step25, r.step24, r.step23, r.step22, r.step21,
		r.step20, r.step19, r.step18, r.step17, r.step16,
		r.step15, r.step14, r.step13, r.step12, r.step11,
		r.step10, r.step9, r.step8, r.step7, r.step6,
		r.step5, r.step4, r.step3, r.step2, r.step1,
	}

	for _, s := range steps {
		if err := s.Revert(ctx); err != nil {
			return err
		}
	}

	return nil
}
