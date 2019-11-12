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

set -eu

# Go to root dir
cd "${0%/*}"/..

START_TIME=$(date +%s)
SKIP_INTEGRATION=true ./hack/test.sh $@
END_TIME=$(date +%s)
echo "Unit tests took $(($END_TIME - $START_TIME)) seconds to complete."