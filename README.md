[![Build Status][travisci-image]][travisci-url]
[![Go Report Card][goreport-image]][goreport-url]

# Cherry

This is a **WORK-IN-PROGRESS**.

Cherry is an **opinionated** tool for *testing*, *buidling*, *releasing*, and *deploying* applications.
Currently, Cherry only supports [Go](https://golang.org) applications and [GitHub](https://github.com) repositories.

## Prerequisites

You need to have the following tools installed and ready.

  * [go](https://golang.org)
  * [git](https://git-scm.com)
  * [github_changelog_generator](https://github.com/github-changelog-generator/github-changelog-generator)

For releasing GitHub repository you need a **personal access token** with **admin** access to your repo.

## Quick Start

You can run `cherry` or `cherry -help` to see the list of available commands.
For each command you can then use `-help` flag too see the help text for the command.

### Test

`cherry test` will run your tests the same way as `go test` and generates an **aggregated coverage report**. 

### Build

`cherry build` will compile your binary and injects the build information into the `version` package.
`cherry build -cross-compile` will build the binaries for all supported platforms.

### Release

`cherry release` can be used for releasing a **GitHub** repository.
You can use `-patch`, `-minor`, or `-major` flags to release at different levels.
You can also use `-comment` flag to include a description for your release.
`CHERRY_GITHUB_TOKEN` environment variable should be set to a **personal access token** with **admin** permission to your repo.

## Examples

You can take a look at [examples](./examples) to see how you can use and configure Cherry.

## Development

| Command           | Description                                          |
|-------------------|------------------------------------------------------|
| `make run`        | Run the application locally                          |
| `make build`      | Build the binary locally                             |
| `make build-all`  | Build the binary locally for all supported platforms |
| `make test`       | Run the unit tests                                   |
| `make test-short` | Run the unit tests using `-short` flag               |
| `make coverage`   | Run the unit tests with coverage report              |
| `make docker`     | Build Docker image                                   |
| `make push`       | Push built image to registry                         |


[travisci-url]: https://travis-ci.org/moorara/cherry
[travisci-image]: https://travis-ci.org/moorara/cherry.svg?branch=master

[goreport-url]: https://goreportcard.com/report/github.com/moorara/cherry
[goreport-image]: https://goreportcard.com/badge/github.com/moorara/cherry
