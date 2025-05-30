
mockery
=======
[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/vektra/mockery/v3/template) ![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/vektra/mockery) ![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/vektra/mockery) [![Go Report Card](https://goreportcard.com/badge/github.com/vektra/mockery/v3)](https://goreportcard.com/report/github.com/vektra/mockery/v3) [![codecov](https://codecov.io/gh/vektra/mockery/branch/master/graph/badge.svg)](https://codecov.io/gh/vektra/mockery)

mockery provides the ability to easily generate mocks for Golang interfaces using the [stretchr/testify/mock](https://pkg.go.dev/github.com/stretchr/testify/mock?tab=doc) package. It removes the boilerplate coding required to use mocks.

Documentation
--------------

Documentation is found at our [GitHub Pages site](https://vektra.github.io/mockery/).

Development
------------

taskfile.dev is used for build tasks. Initialize all go build tools:

```
go mod download -x
```

You can run any of the steps listed in `Taskfile.yml`:

```
$ task test
task: [test] go test -v -coverprofile=coverage.txt ./...
```

Stargazers
----------

[![Stargazers over time](https://starchart.cc/vektra/mockery.svg)](https://starchart.cc/vektra/mockery)
