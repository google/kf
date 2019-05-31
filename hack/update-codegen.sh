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

ROOT_PACKAGE="github.com/GoogleCloudPlatform/kf"
CUSTOM_RESOURCE_NAME="kf"
CUSTOM_RESOURCE_VERSION="v1alpha1"

export GO111MODULE=off

if [ "$(realpath $(go env GOPATH)/src/${ROOT_PACKAGE})" != "$(git rev-parse --show-toplevel)" ]; then
    echo "The generator scripts aren't go module compatible (yet)."
    exit 1
fi

root_dir=$(git rev-parse --show-toplevel)

# retrieve the code-generator scripts and bins
go get -u k8s.io/code-generator/...
cd $(go env GOPATH)/src/k8s.io/code-generator

# run the code-generator entrypoint script
./generate-groups.sh all "$ROOT_PACKAGE/pkg/client" "$ROOT_PACKAGE/pkg/apis" "$CUSTOM_RESOURCE_NAME:$CUSTOM_RESOURCE_VERSION" --go-header-file="${root_dir}/pkg/kf/internal/tools/option-builder/LICENSE_HEADER.go.txt"

# Fix issues due to using old k8s.io/client-go
# The generator wants to use a codec that is only available in a version of
# k8s.io/apimachinery that we can't yet use.
os_friendly_sed () {
    sed_args=()
    if [[ "$OSTYPE" == "darwin"* ]]; then
        sed_args+=('-i')
        sed_args+=('')
    fi

    sed_args+=("$1")
    sed_args+=("$2")
    sed "${sed_args[@]}"
}

os_friendly_sed 's/scheme.Codecs.WithoutConversion()/scheme.Codecs/g' "$(go env GOPATH)/src/${ROOT_PACKAGE}/pkg/client/clientset/versioned/typed/kf/v1alpha1/kf_client.go"
os_friendly_sed 's/pt, //g' "$(go env GOPATH)/src/${ROOT_PACKAGE}/pkg/client/clientset/versioned/typed/kf/v1alpha1/fake/fake_commandset.go"
