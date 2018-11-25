[![Build Status][travisci-image]][travisci-url]
[![Go Report Card][goreport-image]][goreport-url]

# Cherry

This is a **work-in-progress** opinionated tool for buidling, releasing, and deploying applications.

## Commands

| Command                        | Description                                          |
|--------------------------------|------------------------------------------------------|
| `make run`                     | Run the application locally                          |
| `make build`                   | Build the binary locally                             |
| `make build-all`               | Build the binary locally for all supported platforms |
| `make test`                    | Run the unit tests                                   |
| `make test-short`              | Run the unit tests using `-short` flag               |
| `make coverage`                | Run the unit tests with coverage report              |
| `make docker`                  | Build Docker image                                   |
| `make push`                    | Push built image to registry                         |


[travisci-url]: https://travis-ci.org/moorara/cherry
[travisci-image]: https://travis-ci.org/moorara/cherry.svg?branch=master

[goreport-url]: https://goreportcard.com/report/github.com/moorara/cherry
[goreport-image]: https://goreportcard.com/badge/github.com/moorara/cherry
