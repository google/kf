# Assorted scripts for development

This directory contains several scripts useful in the development process of
kf.

- `build.sh` Build the kf CLI.
- `check-go-fmt.sh` Runs the Go formatter and checks to see if anything
  changed. This is useful for enforcing commits be formatted.
- `checks-go-generate.sh` Runs `go generate ./...` and checks to see if
  anything changed. This is useful for enforcing commits be formatted.
- `clean-integration-tests.sh` Deletes apps that might have been left over
  from running the integration tests.
- `test.sh` Run all the tests including the integration tests.
- `update-codegen.sh` Updates auto-generated client libraries.
