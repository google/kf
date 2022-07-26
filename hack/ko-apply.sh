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

cd "${0%/*}"/..

source ./hack/util.sh

tempfile="$(mktemp)"
trap 'rm -f $tempfile' EXIT

VERSION="${VERSION:-$(version)}"

# This is necessary as of ko v0.6.0. They updated the default base image to
# use nonroot and therefore broke our Build pipeline which currently requires
# root. This may not be necessary after b/160025670 is fixed.
export KO_DEFAULTBASEIMAGE="gcr.io/distroless/static:latest"

export KO_DOCKER_REPO="${KO_DOCKER_REPO:-gcr.io/$(gcloud config get-value project)}"

./hack/generate-lifecycle-artifacts.sh

ko resolve --filename config | sed "s/VERSION_PLACEHOLDER/${VERSION}/" >"$tempfile"
kubectl apply -f "$tempfile"
