#!/usr/bin/env bash

# Copyright 2020 Google LLC
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
cd "${0%/*}"/..

# Selecting the correct version of Istio should be done by looking at
# what version the ASM is using:
# https://cloud.google.com/service-mesh/docs/gke-install-asm
#
# NOTE: You might have to look around and find the curl that referencing what
# version is being used under the hood.

tag=$1
if [ -z "${tag}" ]; then
    echo "Usage: $0 [GIT TAG]"
    exit 1
fi

set -u

# Look for annotated tag first. If we don't find that, try lightweight one.
tag_patterns=()
# Annotated
tag_patterns+=("refs/tags/${tag}\^\{\}$")
# Lightweight
tag_patterns+=("refs/tags/${tag}$")

sha=""
for tag_pattern in "${tag_patterns[@]}"; do
    sha=$(git ls-remote https://github.com/istio/api | grep -E "${tag_pattern}" | awk '{print $1}')
    if [ -n "${sha}" ]; then
        break
    fi
done

if [ -z "${sha}" ]; then
    echo "Unable to find tag ${tag}"
    exit 1
fi

go get "istio.io/api@${sha}"

# Update replace directive
mod_version=$(go mod edit --json | jq -r '.Require[] | select(.Path=="istio.io/api") | .Version')
go mod edit --replace="istio.io/api=istio.io/api@${mod_version}"
