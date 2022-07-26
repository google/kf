#!/bin/bash
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

export GO111MODULE=on
export GOFLAGS="-mod=vendor"

REPO_ROOT=$(dirname "${BASH_SOURCE}")/..

# Ensure all go files are formatted
go_changes=$(git status --porcelain | grep .go || true)
if [ -n "${go_changes}" ]; then
   echo "ERROR: This CL contains misformatted golang files. To fix this error, run `gofmt -w -s` on the affected files and add the updated files to this CL."
   echo "stale files:"
   printf "${go_changes}\n"
   echo "git diff:"
   git --no-pager diff
   exit 1
fi

function run_golint_check() {
   local golint_result=""
   for pkg in "$@"
   do
      golint_result+=$(golint "$pkg")
   done

   if [ -n "$golint_result" ]; then
      echo "ERROR: This CL contains golint errors."
      echo "$golint_result"
      return 1
   fi

   return 0
}

# Ensure no golint errors.
run_golint_check "cmd/..." \
   "pkg/apis/..." \
   "pkg/manifestival/..." \
   "pkg/operand/..." \
   "pkg/operator-cleanup/..." \
   "pkg/reconciler/..." \
   "pkg/release/..." \
   "pkg/testing/..." \
   "pkg/transformer/..." \
   "version/..."

TMP="$(mktemp -d)"
cp -r ${REPO_ROOT}/* ${TMP}/
pushd "${TMP}"

# Perform go build
go build -v ./...

popd
