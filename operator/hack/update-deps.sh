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


set -exuo pipefail

export GO111MODULE=on
export GOFLAGS=""
# Required to vendor the entitlement library
export GOPRIVATE='*.googlesource.com,*.git.corp.google.com'

source "$(dirname "${BASH_SOURCE[0]}")/../vendor/knative.dev/hack/library.sh"
source "${REPO_ROOT_DIR}/hack/library.sh"

cd ${REPO_ROOT_DIR}

# Used to pin floating deps to a release version.
VERSION="$(cat hack/VERSION)"
# The list of dependencies that we track at HEAD and periodically
# float forward in this repository.
FLOATING_DEPS=(
  "knative.dev/operator@${VERSION}"
  "knative.dev/pkg@${VERSION}"
  "knative.dev/serving@${VERSION}"
  "knative.dev/hack@${VERSION}"
)

# Parse flags to determine any we should pass to dep.
GO_GET=0
while [[ $# -ne 0 ]]; do
  parameter=$1
  case ${parameter} in
    --upgrade) GO_GET=1 ;;
    *) abort "unknown option ${parameter}" ;;
  esac
  shift
done
readonly GO_GET

if (( GO_GET )); then
  warn_if_not_update_vendors
  go get -d "${FLOATING_DEPS[@]}"
fi

# Prune modules.
echo "Running go mod..."
go mod vendor

export GOFLAGS=-mod=vendor
echo "Removing owner files..."
find vendor -name OWNERS -delete

remove_broken_symlinks ./vendor

echo "Updating vendor licenses..."
TEMP=$(mktemp -t cloudrun-operator-gather-go-licenses-XXXXXXXX.txt)
trap "rm $TEMP" EXIT

# This VENDOR-LICENSE is included in the operator's container image via
# a symlink from the kodata directory. This command also has expected failures
# due to "gke-internal" dependencies not have licenses, but other licenses
# will be updated.
update_licenses third_party/VENDOR-LICENSE "./..." || true

# For now follow the previous pattern and exclude gke-internal from the
# VENDOR-LICENSE directory before packaging.  This isn't vendor code,
# so doesn't belong here.
rm -rf "${REPO_ROOT_DIR}/third_party/VENDOR-LICENSE/gke-internal"

set +x
echo "Please ignore error message about one or more libraries have an incompatible/unknown license (ends with --- FAIL: go-licenses failed to update licenses), those are expected failures."

echo "--- PASS: update-deps complete"
