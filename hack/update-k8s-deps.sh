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


# Selecting the correct version of Kubernetes should be done by looking at
# what version the regular GKE release channel is using:
# https://cloud.google.com/kubernetes-engine/docs/release-notes-regular

set -ex
cd "${0%/*}"/..

k8s_version=$1
if [ -z "${k8s_version}" ]; then
    echo "Usage: $0 [K8s Version]"
    exit 1
fi

function fetch_sha() {
    repo=$1
    tag=$2

    # Look for annotated tag first. If we don't find that, try lightweight one.
    tag_patterns=()
    # Annotated
    tag_patterns+=("refs/tags/${tag}\^\{\}$")
    # Lightweight
    tag_patterns+=("refs/tags/${tag}$")

    sha=""
    for tag_pattern in "${tag_patterns[@]}"; do
        sha="$(git ls-remote "https://github.com/kubernetes/${repo}" | grep -E "${tag_pattern}" | awk '{print $1}')"
        if [ -n "${sha}" ]; then
            break
        fi
    done

    if [ -z "${sha}" ]; then
        echo "Unable to find tag ${tag} in ${repo}"
        exit 1
    fi

    echo "${sha}"
}

temp_dir=$(mktemp -d)
git clone \
    --depth=1 https://github.com/kubernetes/kubernetes \
    --branch "${k8s_version}" \
    --single-branch \
    "${temp_dir}/kubernetes"

staged_repos=()
pushd "${temp_dir}/kubernetes/staging/src/k8s.io"
    # sha=$(fetch_sha "kubernetes" ${k8s_version})
    # git checkout ${sha}

    for dir in ./sample-*; do
        staged_repos+=("$dir")
    done
popd

# Setup the tag from the version. The librarires have tags of the form
# kubernetes-{Major}.{Minor}.{Patch}
tag="kubernetes-$(echo "${k8s_version}" | cut -d v -f2)"

for repo in "${staged_repos[@]}"; do
    echo "Looking for github.com/kubernetes/${repo}"
    sha="$(fetch_sha "${repo}" "${tag}")"

    go mod edit --dropreplace="k8s.io/${repo}" --droprequire="k8s.io/${repo}"
    go get -d "k8s.io/${repo}@${sha}"

    # Update replace directive
    mod_version=$(go mod edit --json | jq -r ".Require[] | select(.Path==\"k8s.io/${repo}\") | .Version")

    if [ -n "${mod_version}" ]; then
        go mod edit --replace="k8s.io/${repo}=k8s.io/${repo}@${mod_version}"
    fi
done
