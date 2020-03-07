# Contributing to Granitic

We are always delighted to receive contributions from the community. Please follow the rules below.

1. You must abide by the [code of conduct](CODE_OF_CONDUCT.md).
1. Please only submit work for an [open issue](https://github.com/graniticio/granitic/issues) - if you have found a bug 
or are proposing an enhancement, please open a new issue tagged with an appropriate [milestone](https://github.com/graniticio/granitic/milestones).
1. If you are unsure about any part of the issue, please submit a comment asking for clarification.
1. Please be a confident Go developer and Git/GitHub user. 
1. Follow the design guidelines below.
1. You must never introduce a dependency on an external package, even for testing.
1. You must follow the backwards compatibility policy below.
1. If your contribution extends or alters functionality that is documented in the [reference manual](https://granitic.io/ref/) 
you must update the reference manual Markdown files (found in the `doc/ref` folder).
1. Make sure that you can run `go test ./...` without failures before submitting. See the test guidelines below.
1. Make sure you have run `go fmt` on all Go source files.
1. Make sure that you can run `go vet ./...` and `go lint ./...` without any warnings.
1. Submit your contribution as a [pull request](https://help.github.com/en/github/collaborating-with-issues-and-pull-requests/about-pull-requests).
1. Choose the correct target branch for your contribution. See the branch policy below.

## Design guidelines

Granitic is a framework based around the concepts of configurability and inversion as control. When contributing,
please understand the role of the code you are modifying. If it is part of a [Granitic facility](https://granitic.io/ref/facilities)
you will generally want to make sure any additional configurability is managed through the configuration files 
(in `facility/config`) and builders (under the `facility` package).

Because of the dynamic nature of type instantiation in Granitic and the passing round of component references, 
you will see more uses of struct pointers and `new()` than is typical for Go applications. This is normally
by design, so be very careful if you refactor types.

If you find code that is in obvious need of a major refactor, it is best to raise an issue against an upcoming
_major_ release than doing the work in a maintenance or minor release branch.

## Compatibility

Granitic's rules for backwards compatibility are straightforward - applications developed with a particular
version of Granitic are guaranteed to build and run against with any later version of Granitic until the
major version number changes.

So an application built against version `2.0.3` will work with `2.1.0`, `2.2.5` etc, but is not guaranteed to build against `3.0.0` or later.

You must respect this promise with any contribution you make.


## Branches

Granitic runs two types of long-term Git branch. Branches of the form `x.y-maint` (e.g `2.1-maint`) are maintenance
branches for major and minor versions of Granitic. If your issue is a bug fix, you should target your pull request 
to one of these branches.

The other type of branch is of the form `dev-x.y` (e.g `dev-2.2`) which is the development branch for unreleased 
features.

Generally you can tell which branch you should be submitting to based on the [milestone](https://github.com/graniticio/granitic/milestones)
associated with the issue.


## Tests

Please include high-quality tests with your submissions. The rule with test coverage is that your contribution 
must not lower the coverage percentage of the file or package you are modifying.

If you are creating a new Go file as part of your submission, the expectation is that your tests will provide
100% coverage for that file.

Granitic's strict no-dependencies rule extends to tests. You can only use the test libraries included as part of
Go or in Granitic's `test` package.

If your tests require data files, they should be included in a folder named `testdata` in the package you are testing.