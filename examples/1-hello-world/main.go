package main

import (
	"fmt"

	"github.com/moorara/cherry/examples/1-hello-world/version"
)

func main() {
	fmt.Printf("Version:   %s\n", version.Version)
	fmt.Printf("Commit:    %s\n", version.Commit)
	fmt.Printf("Branch:    %s\n", version.Branch)
	fmt.Printf("GoVersion: %s\n", version.GoVersion)
	fmt.Printf("BuildTool: %s\n", version.BuildTool)
	fmt.Printf("BuildTime: %s\n", version.BuildTime)
}
