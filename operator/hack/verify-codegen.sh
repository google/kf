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


set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

export GO111MODULE=on

REPO_ROOT_DIR="$(dirname $0)/.."
GKEINTERNAL_DIR="${REPO_ROOT_DIR}/../.."

# Make sure our dependencies are up-to-date
"${REPO_ROOT_DIR}/hack/update-codegen.sh"

git add -A
vendor_diffs="$(git diff HEAD)"
if [[ -n "${vendor_diffs}" ]]; then
    echo "ERROR: generated code not up-to-date, please run ./hack/update-codegen.sh"
    echo "Git diff:"
    echo ${vendor_diffs}
    exit 1
fi

echo "Generated code up-to-date"
exit 0
