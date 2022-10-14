#!/usr/bin/env bash

# Copyright 2019 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the License);
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an AS IS BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -e

cd "${0%/*}"/..

echo "Executing integration tests from: $(pwd)"
export KUBECTL=${KUBECTL:-kubectl}

temp_dir=$(mktemp -d)
trap 'rm -rf ${temp_dir}' EXIT

go clean -testcache

# We need the retry function.
source ./hack/util.sh

export APP_CACHE=${temp_dir}/images.json
export RANDOM_SPACE_NAMES=true
export SPACE_DOMAIN="${TEST_DOMAIN:-integration-tests.kf.dev}"
echo "Space domain: ${SPACE_DOMAIN}"
${KUBECTL} delete spaces --all

START_TIME=$(date +%s)

# When running integration tests, we would like two things that simply running
# go test -v ./... doesn't provide:
# 1. Don't run each package in parallel - This puts quite a bit of burden on
# the cluster and causes tests to flake.
# 2. We want logs streaming the whole time, not just when the tests fails.
#
# Note: We also enforce that all integration tests have the naming convention
# TestIntegration_XXX. Therefore we can single them out.
# NOTE: The tests in ./pkg/kf/internal/last_integration_tests are ran last.

# Find all the integration tests and put
# ./pkg/kf/internal/last_integration_tests/integration_test.go at the end.
arr=(`find . -name "integration_test.go" | grep -v "./pkg/kf/internal/last_integration_tests/integration_test.go"`)
arr+=("./pkg/kf/internal/last_integration_tests/integration_test.go")

for f in "${arr[@]}"; do
  echo "Executing sub-integration tests for: $(dirname "$f")"
  START_SUBTEST=$(date +%s)
  GOMAXPROCS=24 go run ./cmd/test-runner --attempts=3 --timeout=60m --run=TestIntegration_ $(dirname "$f")
  END_SUBTEST=$(date +%s)
  echo "Sub-integration tests for $(dirname "$f") took $((END_SUBTEST - START_SUBTEST)) seconds to complete."
done

END_TIME=$(date +%s)
echo "Integration tests took $((END_TIME - START_TIME)) seconds to complete."
