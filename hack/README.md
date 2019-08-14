# Assorted scripts for development

This directory contains several scripts useful in the development process of
kf.

- `build.sh` Build the kf CLI.
- `build-release.sh` Build the release and CLI artifacts.
- `checks-go-generate.sh` Runs `go generate ./...` and checks to see if
  anything changed. This is useful for enforcing commits be formatted.
- `check-go-sum.sh` Ensures the `go.sum` file is correct.
- `check-linters.sh` Runs the linters and checks to see if anything is wrong.
- `clean-integration-tests.sh` Deletes apps that might have been left over
  from running the integration tests.
- `test.sh` Run all the tests including the integration tests.
- `update-codegen.sh` Updates auto-generated client libraries.
- `upload-release.sh` Builds and uploads the release YAML and CLI artifacts
  for the nightly builds.
