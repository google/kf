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

if [ "${SKIP_INTEGRATION:-false}" = "true" ]; then
    echo "SKIP_INTEGRATION set to 'true'. Skipping integration tests..."
else
  if [ "${GCP_PROJECT_ID}" = "" ]; then
    export GCP_PROJECT_ID=$(gcloud config get-value project)
  fi
fi

green() {
    echo -e "\033[32m$1\033[0m"
}

red() {
    echo -e "\033[31m$1\033[0m"
}

args="-v"
if [ ! "${NO_RACE:-false}" = "true" ]; then
  echo disabling race
  args="--race $args"
fi

go test $args ./...
ret=$?
set +x
if [ $ret -eq 0 ]; then
  green Success
  exit 0
else
  red Failure
  exit $ret
fi
