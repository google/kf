---
title: "Testing Kf Components"
weight: 500
description: Understand the Kf test suite and what it covers.
---

This page describes the tests that Kf runs and notable gaps you may want to cover
when using Kf as your platform.

## Components

### Unit tests

Kf uses Go's unit testing functionality to perform extensive validation of the business logic
in the CLI, the control plane components, and "golden" files which validate the structure of I/O
that Kf expects to be deterministic (e.g. certain CLI output).

Unit tests can be executed using the `./hack/unit-test.sh` shell script in the Kf repository.
This script skips exercising the end to end tests. 

Unit tests are naturally limited because they have to mock interactions with external
services e.g. the Kubernetes control plane and container registries.

### End to end tests

Kf uses specially marked Go tests for acceptance and end to end testing.

These can be executed using the `./hack/integration-test.sh` and `./hack/acceptance-test.sh` 
shell scripts. These scripts will test against the currently targeted Kf cluster.

These tests check for behavior between Kf and other components, but are still limited because
they tend to test one specific thing at a time and are built for speed. 

### Load tests

Kf can run load tests against the platform using the `./ci/cloudbuild/test.yaml` Cloud Build
template with the `_STRESS_TEST_BASELOAD_FACTOR` environment variable set.

This will deploy the Diego stresds test workloads to Kf to check how the platform behaves
with failing Apps.

## Gaps

Kf runs unit tests and end to end tests for each release. You may want to augment with 
additional qualificatoin tests on your own cluster to check for:

* Compatibility with your service brokers.
* Compatibility with all buildpacks you normally use.
* Compatibility with represenative workloads.
* Compatibility with intricate combinations of features (e.g. service bindings, docker containers, buildpacks, routing).
* Compatibility with non-Kf Kubernetes components.