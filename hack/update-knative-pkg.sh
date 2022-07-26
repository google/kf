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


# Selecting the correct version of Knative should be done by looking at what
# Tekton uses.
# NOTE: github.com/knative/pkg uses branches of the form
# release-{Major}.{Minor}.

set -e
cd "${0%/*}"/..

branch=$1
if [ -z "${branch}" ]; then
    echo "Usage: $0 [GIT BRANCH]"
    exit 1
fi

set -u

sha="$(git ls-remote https://github.com/knative/pkg | grep -E "refs/heads/${branch}" | awk '{print $1}')"

if [ -z "${sha}" ]; then
    echo "Unable to find branch ${branch}"
    exit 1
fi

go get "knative.dev/pkg@${sha}"

# Update replace directive
mod_version=$(go mod edit --json | jq -r '.Require[] | select(.Path=="knative.dev/pkg") | .Version')
go mod edit --replace="knative.dev/pkg=knative.dev/pkg@${mod_version}"
