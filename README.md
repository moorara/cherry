[![Build Status][workflow-image]][workflow-url]
[![Go Report Card][goreport-image]][goreport-url]
[![Test Coverage][coverage-image]][coverage-url]
[![Maintainability][maintainability-image]][maintainability-url]

# Cherry

Cherry is an experimental tool and it is **WORK-IN-PROGRESS**.

Cherry is an **opinionated** tool for _buidling_ and _releasing_ applications.
Currently, Cherry only supports [Go](https://golang.org) for building and [GitHub](https://github.com) repositories for releasing.

For Go applications, Cherry supports cross-compiling and injecting metadata into the binaries.

## Prerequisites/Dependencies

You need to have the following tools installed and ready.

  * [git](https://git-scm.com)
  * [go](https://golang.org)
  * [github_changelog_generator](https://github.com/github-changelog-generator/github-changelog-generator)

For releasing GitHub repository you need a **personal access token** with **admin** access to your repo.

## Quick Start

### Install

```
curl -s https://git.io/JeCX6 | sh
```

Or

```
curl -s https://raw.githubusercontent.com/moorara/cherry/master/scripts/install.sh | sh
```

### Docker

The docker image for Cherry includes all the required tools and is accessible at [moorara/cherry](https://hub.docker.com/r/moorara/cherry).

### Examples

You can take a look at [examples](./examples) to see how you can use and configure Cherry.

## Versioning

Cherry uses Semantic Versioning 2.0.0 as described [here](https://semver.org).
For injecting versioning and other metadata into your binaries,
you need to have a `version` package with the following exported variables in it:

```go
var (
  Version   string
  Commit    string
  Branch    string
  GoVersion string
  BuildTool string
  BuildTime string
)
```

The initial release is always `0.1.0`.

## Commands

You can run `cherry` or `cherry -help` to see the list of available commands.
For each command you can then use `-help` flag too see the help text for the command.

### Build

`cherry build` will compile your binary and injects the build information into the `version` package.
`cherry build -cross-compile` will build the binaries for all supported platforms.

### Release

`cherry release` can be used for releasing a **GitHub** repository.
You can use `-patch`, `-minor`, or `-major` flags to release at different levels.
You can also use `-comment` flag to include a description for your release.

`CHERRY_GITHUB_TOKEN` environment variable should be set to a **personal access token** with **admin** permission to your repo.

### Update

`cherry update` will update Cherry to the latest version.
It downloads the latest release for your system from GitHub and replaces the local binary.

## Development

| Command            | Description                                          |
|--------------------|------------------------------------------------------|
| `make build`       | Build the binary locally                             |
| `make build-all`   | Build the binary locally for all supported platforms |
| `make test`        | Run the unit tests                                   |
| `make test-short`  | Run the unit tests using `-short` flag               |
| `make coverage`    | Run the unit tests with coverage report              |
| `make docker`      | Build Docker image                                   |
| `make push`        | Push built image to registry                         |
| `make save-docker` | Save built image to disk                             |
| `make load-docker` | Load saved image from disk                           |


[workflow-url]: https://github.com/moorara/cherry/actions
[workflow-image]: https://github.com/moorara/cherry/workflows/Main/badge.svg
[goreport-url]: https://goreportcard.com/report/github.com/moorara/cherry
[goreport-image]: https://goreportcard.com/badge/github.com/moorara/cherry
[coverage-url]: https://codeclimate.com/github/moorara/cherry/test_coverage
[coverage-image]: https://api.codeclimate.com/v1/badges/569a659577775c8af668/test_coverage
[maintainability-url]: https://codeclimate.com/github/moorara/cherry/maintainability
[maintainability-image]: https://api.codeclimate.com/v1/badges/569a659577775c8af668/maintainability
