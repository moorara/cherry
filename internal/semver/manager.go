package semver

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	jsonExt  = ".json"
	textFile = "VERSION"
)

type (
	// Manager is the interface for a semantic versioning manager
	Manager interface {
		Read() (SemVer, error)
		Update(string) error
	}

	textManager struct {
		file string
	}

	jsonManager struct {
		file string
	}
)

// NewManager creates a new manager for a version file
func NewManager(file string) (Manager, error) {
	if filepath.Base(file) == textFile {
		return NewTextManager(file), nil
	} else if filepath.Ext(file) == jsonExt {
		return NewJSONManager(file), nil
	}

	return nil, fmt.Errorf("unknown version file: %s", file)
}

// NewTextManager creates a new manager for a text version file
func NewTextManager(file string) Manager {
	return &textManager{
		file: file,
	}
}

func (m *textManager) Read() (SemVer, error) {
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

func (m *textManager) Update(version string) error {
	data := []byte(version + "\n")

	err := ioutil.WriteFile(m.file, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

// NewJSONManager creates a new manager for a json version file (package.json)
func NewJSONManager(file string) Manager {
	return &jsonManager{
		file: file,
	}
}

func (m *jsonManager) Read() (SemVer, error) {
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

func (m *jsonManager) Update(version string) error {
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
