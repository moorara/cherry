package version

import (
	"fmt"
)

var (
	// Version is the semantic version
	Version string

	// Revision is the SHA-1 of the git revision
	Revision string

	// Branch is the name of the git branch
	Branch string

	// GoVersion is the go compiler version
	GoVersion string

	// BuildTool contains the name and version of build tool
	BuildTool string

	// BuildTime is the time binary built
	BuildTime string
)

// Get returns a string describing the version in details
func Get() string {
	return fmt.Sprintf("%s %s %s %s %s", Version, Revision, Branch, GoVersion, BuildTime)
}
