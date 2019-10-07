package step

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/moorara/cherry/pkg/semver"
)

const (
	textFile = "VERSION"
	jsonFile = "package.json"
)

func findVersionFile(workDir, filename string) string {
	if filename != "" {
		return filename
	} else if _, err := os.Stat(filepath.Join(workDir, textFile)); err == nil {
		return textFile
	} else if _, err := os.Stat(filepath.Join(workDir, jsonFile)); err == nil {
		return jsonFile
	}

	return ""
}

// SemVerRead reads version from version file.
type SemVerRead struct {
	Mock     Step
	WorkDir  string
	Filename string
	Result   struct {
		Filename string
		Version  semver.SemVer
	}
}

func (s *SemVerRead) readVersion() (string, semver.SemVer, error) {
	var zero semver.SemVer
	var versionFile, versionString string

	if versionFile = findVersionFile(s.WorkDir, s.Filename); versionFile == "" {
		return "", zero, errors.New("no version file")
	}

	versionFilePath := filepath.Join(s.WorkDir, versionFile)
	data, err := ioutil.ReadFile(versionFilePath)
	if err != nil {
		return "", zero, err
	}

	if filepath.Ext(versionFilePath) == ".json" { // package.json file
		packageJSON := struct {
			Version string `json:"version"`
		}{}

		err = json.Unmarshal(data, &packageJSON)
		if err != nil {
			return "", zero, err
		}

		versionString = packageJSON.Version
	} else { // text file
		versionString = strings.Trim(string(data), "\n")
	}

	if versionString == "" {
		return "", zero, errors.New("empty version")
	}

	version, err := semver.Parse(versionString)
	if err != nil {
		return "", zero, err
	}

	return versionFile, version, nil
}

// Dry is a dry run of the step.
func (s *SemVerRead) Dry(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Dry(ctx)
	}

	_, _, err := s.readVersion()
	if err != nil {
		return err
	}

	return nil
}

// Run executes the step.
func (s *SemVerRead) Run(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Run(ctx)
	}

	filename, version, err := s.readVersion()
	if err != nil {
		return err
	}

	s.Result.Filename = filename
	s.Result.Version = version

	return nil
}

// Revert reverts back an executed step.
func (s *SemVerRead) Revert(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Revert(ctx)
	}

	return nil
}

// SemVerUpdate writes a version to version file.
type SemVerUpdate struct {
	Mock     Step
	WorkDir  string
	Filename string
	Version  string
	Result   struct {
		Filename string
	}
}

func (s *SemVerUpdate) writeVersion(dryRun bool) (string, error) {
	var versionFile, content string

	if versionFile = findVersionFile(s.WorkDir, s.Filename); versionFile == "" {
		return "", errors.New("no version file")
	}

	versionFilePath := filepath.Join(s.WorkDir, versionFile)
	if _, err := os.Stat(versionFilePath); os.IsNotExist(err) {
		return "", errors.New("version file not found")
	}

	if filepath.Ext(versionFilePath) == ".json" { // package.json file
		data, err := ioutil.ReadFile(versionFilePath)
		if err != nil {
			return "", err
		}

		re := regexp.MustCompile(`"version":\s*"[^"]*"`)
		content = re.ReplaceAllLiteralString(string(data), fmt.Sprintf(`"version": "%s"`, s.Version))
	} else { // text file
		content = s.Version + "\n"
	}

	if !dryRun {
		err := ioutil.WriteFile(versionFilePath, []byte(content), 0644)
		if err != nil {
			return "", err
		}
	}

	return versionFile, nil
}

// Dry is a dry run of the step.
func (s *SemVerUpdate) Dry(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Dry(ctx)
	}

	filename, err := s.writeVersion(true)
	if err != nil {
		return err
	}

	s.Result.Filename = filename

	return nil
}

// Run executes the step.
func (s *SemVerUpdate) Run(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Run(ctx)
	}

	filename, err := s.writeVersion(false)
	if err != nil {
		return err
	}

	s.Result.Filename = filename

	return nil
}

// Revert reverts back an executed step.
func (s *SemVerUpdate) Revert(ctx context.Context) error {
	if s.Mock != nil {
		return s.Mock.Revert(ctx)
	}

	return nil
}
