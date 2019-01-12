package main

import (
	"fmt"

	"github.com/moorara/cherry/examples/1-hello-world/cmd/version"
)

func main() {
	fmt.Printf(`
	version:    %s
	revision:   %s
	branch:     %s
	goVersion:  %s
	buildTool:  %s
	buildTime:  %s`+"\n\n", version.Version, version.Revision, version.Branch, version.GoVersion, version.BuildTool, version.BuildTime)
}
