#!/usr/bin/env sh

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

# Go to root dir
cd $(git rev-parse --show-toplevel)

if [ "${DOCKER_REGISTRY}" = "" ]; then
  echo running integration tests
  export DOCKER_REGISTRY="gcr.io/$(gcloud config get-value project)"
fi

# When running integration tests, we would like two things that simply running
# go test -v ./... doesn't provide:
# 1. Don't run each package in parallel - This puts quite a bit of burden on
# the cluster and causes tests to flake.
# 2. We want logs streaming the whole time, not just when the tests fails.
#
# Note: We also enforce that all integration tests have the naming convention
# TestIntegration_XXX. Therefore we can single them out.
for f in $(find . | grep integration_test.go); do
  go test -v --timeout=30m $(dirname $f) --run TestIntegration_
done
