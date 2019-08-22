# gomod-collector

`gomod-collector` is a fork of Knative's [dep-collector](https://github.com/knative/test-infra/tree/master/tools/dep-collector)
that works with vendored gomod directories.

## Basic Usage

You can run `gomod-collector` on one or more Go vendor directories, and
it will:

1. Read the `modules.txt` file to identify vendored software packages,
1. Search for licenses for each vendored dependency,
1. Dump a file containing the licenses for each vendored import.

For example (single import path):

```shell
$ gomod-collector .
===========================================================
Import: github.com/kf/...

                                 Apache License
                           Version 2.0, January 2004
                        http://www.apache.org/licenses/
...

```

For example (multiple import paths):

```shell
$ gomod-collector ./ ./cmd/sleeper

===========================================================
Import: github.com/kf/...

                                 Apache License
                           Version 2.0, January 2004
                        http://www.apache.org/licenses/
```

## CSV Usage

You can also run `gomod-collector` in a mode that produces CSV output, including
basic classification of the license.

For example:

```shell
$ gomod-collector -csv .
github.com/google/licenseclassifier,Static,,https://github.com/mattmoor/dep-collector/blob/master/vendor/github.com/google/licenseclassifier/LICENSE,Apache-2.0
github.com/google/licenseclassifier/stringclassifier,Static,,https://github.com/mattmoor/dep-collector/blob/master/vendor/github.com/google/licenseclassifier/stringclassifier/LICENSE,Apache-2.0
github.com/sergi/go-diff,Static,,https://github.com/mattmoor/dep-collector/blob/master/vendor/github.com/sergi/go-diff/LICENSE,MIT

```

The columns here are:

- Import Path,
- How the dependency is linked in (always reports "static"),
- A column for whether any modifications have been made (always empty),
- The URL by which to access the license file (assumes `master`),
- A classification of what license this is
  ([using this](https://github.com/google/licenseclassifier)).

## Check mode

`gomod-collector` also includes a mode that will check for "forbidden" licenses.

> In order to run gomod-collector in this mode, you must first run: go get
> github.com/google/licenseclassifier

For example (failing):

```shell
$ gomod-collector -check ./foo/bar/baz
2018/07/20 22:01:29 Error checking license collection: Errors validating licenses:
Found matching forbidden license in "foo.io/bar/vendor/github.com/BurntSushi/toml":WTFPL
```

For example (passing):

```shell
$ gomod-collector -check .
2018/07/20 22:29:09 No errors found.
```
