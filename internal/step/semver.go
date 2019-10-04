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

	"github.com/moorara/cherry/internal/semver"
)

const (
	textFile = "VERSION"
	jsonFile = "package.json"
)

// SemVerRead reads version from version file.
type SemVerRead struct {
	WorkDir  string
	Filename string
	Result   struct {
		Version semver.SemVer
	}
}

func (s *SemVerRead) readVersion() (semver.SemVer, error) {
	if s.Filename == "" {
		if _, err := os.Stat(filepath.Join(s.WorkDir, textFile)); err == nil {
			s.Filename = textFile
		} else if _, err := os.Stat(filepath.Join(s.WorkDir, jsonFile)); err == nil {
			s.Filename = jsonFile
		} else {
			return semver.SemVer{}, errors.New("no version file")
		}
	}

	versionString := ""
	versionFilePath := filepath.Join(s.WorkDir, s.Filename)
	data, err := ioutil.ReadFile(versionFilePath)
	if err != nil {
		return semver.SemVer{}, err
	}

	if filepath.Ext(versionFilePath) == ".json" { // package.json file
		packageJSON := struct {
			Version string `json:"version"`
		}{}

		err = json.Unmarshal(data, &packageJSON)
		if err != nil {
			return semver.SemVer{}, err
		}

		versionString = packageJSON.Version
	} else { // text file
		versionString = strings.Trim(string(data), "\n")
	}

	if versionString == "" {
		return semver.SemVer{}, errors.New("empty version")
	}

	return semver.Parse(versionString)
}

// Dry is a dry run of the step.
func (s *SemVerRead) Dry(ctx context.Context) error {
	_, err := s.readVersion()
	if err != nil {
		return err
	}

	return nil
}

// Run executes the step.
func (s *SemVerRead) Run(ctx context.Context) error {
	version, err := s.readVersion()
	if err != nil {
		return err
	}

	s.Result.Version = version

	return nil
}

// Revert reverts back an executed step.
func (s *SemVerRead) Revert(ctx context.Context) error {
	return nil
}

// SemVerUpdate writes a version to version file.
type SemVerUpdate struct {
	WorkDir  string
	Filename string
	Version  string
}

func (s *SemVerUpdate) writeVersion(dryRun bool) error {
	if s.Filename == "" {
		if _, err := os.Stat(filepath.Join(s.WorkDir, textFile)); err == nil {
			s.Filename = textFile
		} else if _, err := os.Stat(filepath.Join(s.WorkDir, jsonFile)); err == nil {
			s.Filename = jsonFile
		} else {
			return errors.New("no version file")
		}
	}

	content := ""
	versionFilePath := filepath.Join(s.WorkDir, s.Filename)
	if _, err := os.Stat(versionFilePath); os.IsNotExist(err) {
		return errors.New("version file not found")
	}

	if filepath.Ext(versionFilePath) == ".json" { // package.json file
		data, err := ioutil.ReadFile(versionFilePath)
		if err != nil {
			return err
		}

		re := regexp.MustCompile(`"version":\s*"[^"]*"`)
		content = re.ReplaceAllLiteralString(string(data), fmt.Sprintf(`"version": "%s"`, s.Version))
	} else { // text file
		content = s.Version + "\n"
	}

	if !dryRun {
		err := ioutil.WriteFile(versionFilePath, []byte(content), 0644)
		if err != nil {
			return err
		}
	}

	return nil
}

// Dry is a dry run of the step.
func (s *SemVerUpdate) Dry(ctx context.Context) error {
	return s.writeVersion(true)
}

// Run executes the step.
func (s *SemVerUpdate) Run(ctx context.Context) error {
	return s.writeVersion(false)
}

// Revert reverts back an executed step.
func (s *SemVerUpdate) Revert(ctx context.Context) error {
	return nil
}
