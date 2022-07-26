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


# Selecting the correct version of Tekton should be done by looking at what
# the N-1 version is at:
# https://github.com/tektoncd/pipeline/releases

set -e
cd "${0%/*}"/..

version=$1
if [ -z "${version}" ]; then
    echo "Usage: $0 [VERSION]"
    exit 1
fi

set -u

go get "github.com/tektoncd/pipeline@${version}"

# Update replace directive
mod_version="$(go mod edit --json | jq -r '.Require[] | select(.Path=="github.com/tektoncd/pipeline") | .Version')"
go mod edit --replace="github.com/tektoncd/pipeline=github.com/tektoncd/pipeline@${mod_version}"
