#!/usr/bin/env bash

# Copyright 2019 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# This script is used by the CI to check if 'go generate ./...' is up to date.

set -eux

# Change to the project root directory
cd "${0%/*}"/..

# https://github.com/kubernetes-sigs/kubebuilder/issues/359
# K8s code generation is broken when used in projects that use Go modules.
# Until that is fixed, this script will symlink the project into your GOPATH
# and remove the symlink when it is done. A GOPATH is required.
if [[ -z "$GOPATH" ]]; then
  echo "GOPATH must be set"
  exit 1
fi

# Symlink the project into the GOPATH.
# Required until https://github.com/kubernetes-sigs/kubebuilder/issues/359 is fixed
function finish {
  unlink $GOPATH/src/github.com/google/kf
}
trap finish EXIT
mkdir -p $GOPATH/src/github.com/google
ln -s `pwd` $GOPATH/src/github.com/google/kf

hack/update-codegen.sh

if [ ! -z "$(git status --porcelain)" ]; then
    git status
    echo
    echo "The generated files aren't up to date."
    echo "Update them with the 'hack/update-codegen.sh' command."
    git --no-pager diff
    exit 1
fi
