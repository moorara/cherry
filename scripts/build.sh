#!/bin/sh

#
# USAGE:
#   ./build.sh -a
#   ./build.sh -m main.go -b app
#   ./build.sh --all --main main.go --binary builds/app
#

set -euo pipefail


version=$(cat VERSION)
revision=$(git rev-parse --short HEAD)
branch=$(git rev-parse --abbrev-ref HEAD)
goversion=$(go version | cut -d' ' -f3)
buildtool='script'
buildtime=$(date -u +%Y-%m-%dT%H:%M:%SZ)

version_flag="-X $(go list ./cmd/version).Version=$version"
revision_flag="-X $(go list ./cmd/version).Revision=$revision"
branch_flag="-X $(go list ./cmd/version).Branch=$branch"
goversion_flag="-X $(go list ./cmd/version).GoVersion=$goversion"
buildtool_flag="-X $(go list ./cmd/version).BuildTool=$buildtool"
buildtime_flag="-X $(go list ./cmd/version).BuildTime=$buildtime"
ldflags="$version_flag $revision_flag $branch_flag $goversion_flag $buildtool_flag $buildtime_flag"

platforms="linux-386 linux-amd64 darwin-386 darwin-amd64 windows-386 windows-amd64"


process_args() {
  while [ $# -gt 1 ]; do
    key=$1
    case $key in
      -a|--all)
      all=true
      ;;
      -m|--main)
      main=$2
      shift
      ;;
      -b|--binary)
      binary=$2
      shift
      ;;
    esac
    shift
  done

  all=${all:-false}
  main=${main:-"main.go"}
  binary=${binary:-"build/app"}
}

build_binary() {
  go build \
    -ldflags "$ldflags" \
    -o "$binary" \
    "$main"
}

cross_compile() {
  for platform in $platforms; do
    GOOS=$(echo "$platform" | cut -d- -f1) \
	  GOARCH=$(echo "$platform" | cut -d- -f2) \
	  go build \
      -ldflags "$ldflags" \
      -o "$binary-$platform" \
      "$main"
  done
}


process_args "$@"

if [ "$all" = true ]; then
  cross_compile
else
  build_binary
fi
