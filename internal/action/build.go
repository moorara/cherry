package action

import (
	"context"
	"fmt"
	"time"

	"github.com/moorara/cherry/internal/spec"
	"github.com/moorara/cherry/internal/step"
	"github.com/moorara/cherry/pkg/cui"
)

// build is the action for build command.
type build struct {
	ui    cui.CUI
	spec  *spec.Spec // This needs to be a pointer, so updates made by flag.Parse will be available here
	step1 *step.GoList
	step2 *step.SemVerRead
	step3 *step.GitGetHEAD
	step4 *step.GitGetBranch
	step5 *step.GoVersion
	step6 *step.GoBuild
}

// NewBuild creates an instance of Build action.
func NewBuild(ui cui.CUI, workDir string, s *spec.Spec) Action {
	return &build{
		ui:   ui,
		spec: s,
		step1: &step.GoList{
			WorkDir: workDir,
			Package: s.Build.VersionPackage,
		},
		step2: &step.SemVerRead{
			WorkDir:  workDir,
			Filename: s.VersionFile,
		},
		step3: &step.GitGetHEAD{
			WorkDir: workDir,
		},
		step4: &step.GitGetBranch{
			WorkDir: workDir,
		},
		step5: &step.GoVersion{
			WorkDir: workDir,
		},
		step6: &step.GoBuild{
			WorkDir:    workDir,
			LDFlags:    "TBD",
			MainFile:   s.Build.MainFile,
			BinaryFile: s.Build.BinaryFile,
			Platforms:  nil, // TBD
		},
	}
}

func (b *build) getLDFlags() string {
	buildTool := b.spec.ToolName
	if b.spec.ToolVersion != "" {
		buildTool += "@" + b.spec.ToolVersion
	}

	vPkg := b.step1.Result.PackagePath
	versionFlag := fmt.Sprintf("-X %s.Version=%s", vPkg, b.step2.Result.Version.Version())
	revisionFlag := fmt.Sprintf("-X %s.Revision=%s", vPkg, b.step3.Result.ShortSHA)
	branchFlag := fmt.Sprintf("-X %s.Branch=%s", vPkg, b.step4.Result.Name)
	goVersionFlag := fmt.Sprintf("-X %s.GoVersion=%s", vPkg, b.step5.Result.Version)
	buildToolFlag := fmt.Sprintf("-X %s.BuildTool=%s", vPkg, buildTool)
	buildTimeFlag := fmt.Sprintf("-X %s.BuildTime=%s", vPkg, time.Now().UTC().Format(time.RFC3339Nano))
	ldflags := fmt.Sprintf("%s %s %s %s %s %s", versionFlag, revisionFlag, branchFlag, goVersionFlag, buildToolFlag, buildTimeFlag)

	return ldflags
}

// Dry is a dry run of the action.
func (b *build) Dry(ctx context.Context) error {
	// Step 1 to 5 do NOT have hany side effect
	// Thier results are required by getLDFlags()

	if err := b.step1.Run(ctx); err != nil {
		return err
	}

	if err := b.step2.Run(ctx); err != nil {
		return err
	}

	if err := b.step3.Run(ctx); err != nil {
		return err
	}

	if err := b.step4.Run(ctx); err != nil {
		return err
	}

	if err := b.step5.Run(ctx); err != nil {
		return err
	}

	b.step6.LDFlags = b.getLDFlags()
	if b.spec.Build.CrossCompile {
		b.step6.Platforms = b.spec.Build.Platforms
	}

	if err := b.step6.Dry(ctx); err != nil {
		return err
	}

	return nil
}

// Run executes the action.
func (b *build) Run(ctx context.Context) error {
	if err := b.step1.Run(ctx); err != nil {
		return err
	}

	if err := b.step2.Run(ctx); err != nil {
		return err
	}

	if err := b.step3.Run(ctx); err != nil {
		return err
	}

	if err := b.step4.Run(ctx); err != nil {
		return err
	}

	if err := b.step5.Run(ctx); err != nil {
		return err
	}

	b.step6.LDFlags = b.getLDFlags()
	if b.spec.Build.CrossCompile {
		b.step6.Platforms = b.spec.Build.Platforms
	}

	if err := b.step6.Run(ctx); err != nil {
		return err
	}

	for _, bin := range b.step6.Result.Binaries {
		b.ui.Infof("🍒 %s", bin)
	}

	return nil
}

// Revert reverts back an executed action.
func (b *build) Revert(ctx context.Context) error {
	steps := []step.Step{b.step6, b.step5, b.step4, b.step3, b.step2, b.step1}

	for _, s := range steps {
		if err := s.Revert(ctx); err != nil {
			return err
		}
	}

	return nil
}
