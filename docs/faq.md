---
title: FAQ
---

Frequently Asked Questions
===========================

error: `no go files found in root search path`
---------------------------------------------

When using the `packages` feature, `recursive: true` and you have specified a package that contains no `*.go` files, mockery is unable to determine the on-disk location of the package in order to continue the recursive package search. This appears to be a limitation of the [golang.org/x/tools/go/packages](https://pkg.go.dev/golang.org/x/tools/go/packages) package that is used to parse package metadata.

The solution is to create a `.go` file in the package's path and add a `package [name]` directive at the top. It doesn't matter what the file is called. This allows mockery to properly read package metadata.

[Discussion](https://github.com/vektra/mockery/discussions/636)

internal error: package without types was imported
---------------------------------------------------

[https://github.com/vektra/mockery/issues/475](https://github.com/vektra/mockery/issues/475)

This issue indicates that you have attempted to use package in your dependency tree (whether direct or indirect) that uses Go language semantics that your currently-running Go version does not support. The solution:

1. Update to the latest go version
2. Delete all cached packages with `go clean -modcache`
3. Reinstall mockery

Additionally, this issue only happens when compiling mockery from source, such as with `go install`. Our docs [recommend not to use `go install`](../installation#go-install) as the success of your build depends on the compatibility of your Go version with the semantics in use. You would not encounter this issue if using one of the installation methods that install pre-built binaries, like downloading the `.tar.gz` binaries, or through `brew install`.

Semantic Versioning
-------------------

The mockery project follows the standard Semantic Versioning Semantics. The versioning applies to the following areas:

1. The shape of mocks generated by pre-curated templates.
2. Functions and data provided to templates specified with `#!yaml template: "file://"`.
3. Configuration options.

Mockery is not meant to be used as an imported library. Importing mockery code in external modules is not supported.

Mocking interfaces in `main`
----------------------------

When your interfaces are in the main package, you should supply the `--inpackage` flag.
This will generate mocks in the same package as the target code, avoiding import issues.

mockery fails to run when `MOCKERY_VERSION` environment variable is set
------------------------------------------------------------------------

This issue was first highlighted [in this GitHub issue](https://github.com/vektra/mockery/issues/391).

mockery uses the viper package for configuration mapping and parsing. Viper is set to automatically search for all config variables specified in its config struct. One of the config variables is named `version`, which gets mapped to an environment variable called `MOCKERY_VERSION`. If you set this environment variable, mockery attempts to parse it into the `version` bool config.

This is an adverse effect of how our config parsing is set up. The solution is to rename your environment variable to something other than `MOCKERY_VERSION`.
