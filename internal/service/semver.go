package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const (
	jsonExt  = ".json"
	textFile = "VERSION"
)

type (
	// SemVer represents a semantic versioning
	SemVer struct {
		Major int64
		Minor int64
		Patch int64
	}

	// VersionManager is the interface for a semantic versioning manager
	VersionManager interface {
		Read() (SemVer, error)
		Update(string) error
	}

	textVersionManager struct {
		file string
	}

	jsonVersionManager struct {
		file string
	}
)

// Parse reads a semantic version string and returns a SemVer
func Parse(version string) (semver SemVer, err error) {
	re := regexp.MustCompile("[.+-]")
	comps := re.Split(version, -1)
	if len(comps) < 3 {
		err = errors.New("invalid semantic version")
		return
	}

	// Major version number
	semver.Major, err = strconv.ParseInt(comps[0], 10, 64)
	if err != nil {
		err = fmt.Errorf("invalid major version: %v", err)
		return
	}

	// Minor version number
	semver.Minor, err = strconv.ParseInt(comps[1], 10, 64)
	if err != nil {
		err = fmt.Errorf("invalid minor version: %v", err)
		return
	}

	// Patch version number
	semver.Patch, err = strconv.ParseInt(comps[2], 10, 64)
	if err != nil {
		err = fmt.Errorf("invalid patch version: %v", err)
		return
	}

	return
}

// Version returns a string representation of semantic version
func (v SemVer) Version() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// GitTag returns a string representation of semantic version to be used as a git tag
func (v SemVer) GitTag() string {
	return fmt.Sprintf("v%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// PreRelease returns a string representation of semantic version to be used as a prelease version
func (v SemVer) PreRelease() string {
	return fmt.Sprintf("%d.%d.%d-0", v.Major, v.Minor, v.Patch)
}

// ReleasePatch returns the current and next semantic versions for a patch release
func (v SemVer) ReleasePatch() (current SemVer, next SemVer) {
	current = SemVer{
		Major: v.Major,
		Minor: v.Minor,
		Patch: v.Patch,
	}

	next = SemVer{
		Major: v.Major,
		Minor: v.Minor,
		Patch: v.Patch + 1,
	}

	return
}

// ReleaseMinor returns the current and next semantic versions for a minor release
func (v SemVer) ReleaseMinor() (current SemVer, next SemVer) {
	current = SemVer{
		Major: v.Major,
		Minor: v.Minor + 1,
		Patch: 0,
	}

	next = SemVer{
		Major: v.Major,
		Minor: v.Minor + 1,
		Patch: 1,
	}

	return
}

// ReleaseMajor returns the current and next semantic versions for a major release
func (v SemVer) ReleaseMajor() (current SemVer, next SemVer) {
	current = SemVer{
		Major: v.Major + 1,
		Minor: 0,
		Patch: 0,
	}

	next = SemVer{
		Major: v.Major + 1,
		Minor: 0,
		Patch: 1,
	}

	return
}

// NewVersionManager creates a new instance of version manager
func NewVersionManager(file string) (VersionManager, error) {
	if filepath.Base(file) == textFile {
		return NewTextVersionManager(file), nil
	} else if filepath.Ext(file) == jsonExt {
		return NewJSONVersionManager(file), nil
	}

	return nil, fmt.Errorf("unknown version file: %s", file)
}

// NewTextVersionManager creates a new version manager for a text version file
func NewTextVersionManager(file string) VersionManager {
	return &textVersionManager{
		file: file,
	}
}

func (m *textVersionManager) Read() (SemVer, error) {
	var semver SemVer

	data, err := ioutil.ReadFile(m.file)
	if err != nil {
		return semver, err
	}

	version := strings.Trim(string(data), "\n")
	if version == "" {
		return semver, fmt.Errorf("empty version file")
	}

	semver, err = Parse(version)
	if err != nil {
		return semver, err
	}

	return semver, nil
}

func (m *textVersionManager) Update(version string) error {
	data := []byte(version + "\n")

	err := ioutil.WriteFile(m.file, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

// NewJSONVersionManager creates a new version manager for a json version file (package.json)
func NewJSONVersionManager(file string) VersionManager {
	return &jsonVersionManager{
		file: file,
	}
}

func (m *jsonVersionManager) Read() (SemVer, error) {
	var semver SemVer

	data, err := ioutil.ReadFile(m.file)
	if err != nil {
		return semver, err
	}

	content := make(map[string]interface{})
	err = json.Unmarshal(data, &content)
	if err != nil {
		return semver, err
	}

	val, exist := content["version"]
	if !exist {
		return semver, fmt.Errorf("no version key")
	}

	version, ok := val.(string)
	if !ok {
		return semver, fmt.Errorf("bad version key")
	}

	semver, err = Parse(version)
	if err != nil {
		return semver, err
	}

	return semver, nil
}

func (m *jsonVersionManager) Update(version string) error {
	data, err := ioutil.ReadFile(m.file)
	if err != nil {
		return err
	}

	content := string(data)
	re := regexp.MustCompile(`"version":\s*"[^"]*"`)
	content = re.ReplaceAllLiteralString(content, fmt.Sprintf(`"version": "%s"`, version))
	data = []byte(content)

	err = ioutil.WriteFile(m.file, data, 0644)
	if err != nil {
		return err
	}

	return nil
}
