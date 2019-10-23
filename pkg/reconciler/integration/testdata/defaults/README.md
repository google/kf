This directory contains tests for `defaults_test.go`.

Each file in this directory that ends with `.yaml` will be tested. The files
must follow the structure defined by the `DefaultTest` struct.

Each test is intended to validate the defaulting webhook doesn't introduce new
defaults. This can cause reconciliation to get stuck in an infinite loop where
a resource is created, the defaulter changes it, the reconciler notices a
difference and perpetually attempts to refresh it.

NOTE: All the tests are positive.
