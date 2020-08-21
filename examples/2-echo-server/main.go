package main

import (
	"log"
	"net/http"

	"github.com/moorara/cherry/examples/2-echo-server/handler"
	"github.com/moorara/cherry/examples/2-echo-server/version"
)

func main() {
	log.Printf("version: %s  commit: %s  branch: %s  goVersion: %s  buildTool: %s  buildTime: %s\n",
		version.Version, version.Commit, version.Branch, version.GoVersion, version.BuildTool, version.BuildTime)

	http.HandleFunc("/echo", handler.EchoHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
