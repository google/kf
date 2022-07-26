#!/usr/bin/env bash
# Copyright 2022 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


set -e

# Go to root dir
cd "${0%/*}"/..

source ./hack/util.sh

echo "Executing acceptance tests from: $(pwd)"
go clean -testcache

export APP_CACHE=${temp_dir}/images.json
export RANDOM_SPACE_NAMES=true
export SPACE_DOMAIN="${TEST_DOMAIN:-integration-tests.kf.dev}"

${KUBECTL:-kubectl} delete spaces --all

START_TIME=$(date +%s)
RUN_ACCEPTANCE=true GOMAXPROCS=24 retry "go test -v --timeout=60m --run TestAcceptance_ ./..."
END_TIME=$(date +%s)
echo "Acceptance tests took $((END_TIME - START_TIME)) seconds to complete."
