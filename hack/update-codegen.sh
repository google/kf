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

set -eu

GENERATOR_FLAGS=""

while getopts "v" opt; do
  case $opt in
    v)
      echo "WOW"
      set -x
      GENERATOR_FLAGS="-v 5 ${GENERATOR_FLAGS}"
      ;;
  esac
done

ROOT_PACKAGE="github.com/google/kf"
PACKAGE_LOCATION="$(go env GOPATH)/src/$ROOT_PACKAGE"
CUSTOM_RESOURCE_NAME="kf"
CUSTOM_RESOURCE_VERSION="v1alpha1"

export GO111MODULE=off

if [ ! -d "$PACKAGE_LOCATION" ]; then
  echo "Cannot find go package $ROOT_PACKAGE" 1>&2
  exit 1
fi

if [ "$(realpath $PACKAGE_LOCATION)" != "$(git rev-parse --show-toplevel)" ]; then
    echo "The generator scripts aren't go module compatible (yet)." 1>&2
    exit 1
fi

root_dir=$(git rev-parse --show-toplevel)

if [[ ! -d $(go env GOPATH)/src/k8s.io/code-generator ]] || [[ ! -d $(go env GOPATH)/src/k8s.io/code-generator ]]; then
  echo Some required packages are missing. Run ./hack/setup-codegen.sh first 1>&2
  exit 1
fi

pushd $(go env GOPATH)/src/k8s.io/code-generator

# run the code-generator entrypoint script
./generate-groups.sh all "$ROOT_PACKAGE/pkg/client" "$ROOT_PACKAGE/pkg/apis" "$CUSTOM_RESOURCE_NAME:$CUSTOM_RESOURCE_VERSION" --go-header-file="${root_dir}/pkg/kf/internal/tools/option-builder/LICENSE_HEADER.go.txt" ${GENERATOR_FLAGS}
ret=$?

if [ $ret -ne 0 ]; then
  echo Error running code-generator 1>&2
  exit 1
fi

# Fix issues due to using old k8s.io/client-go
# The generator wants to use a codec that is only available in a version of
# k8s.io/apimachinery that we can't yet use.
os_friendly_sed () {
  echo "Applying $1 to $2"
  sed "$1" "$2" > "$2.new"
  mv "$2.new" "$2"
}

TYPES="kf_client $(ls ${PACKAGE_LOCATION}/pkg/apis/kf/v1alpha1/ | grep 'types.go' | sed 's/_types.go//')"
for type in ${TYPES}; do
  os_friendly_sed 's/scheme.Codecs.WithoutConversion()/scheme.Codecs/g' "${PACKAGE_LOCATION}/pkg/client/clientset/versioned/typed/kf/v1alpha1/${type}.go"
  os_friendly_sed 's/pt, //g' "$(go env GOPATH)/src/${ROOT_PACKAGE}/pkg/client/clientset/versioned/typed/kf/v1alpha1/fake/fake_${type}.go"
done

popd

# Do Knative injection generation
KNATIVE_CODEGEN_PKG=$(go env GOPATH)/src/knative.dev/pkg

${KNATIVE_CODEGEN_PKG}/hack/generate-knative.sh "injection" \
  "github.com/google/kf/pkg/client" "github.com/google/kf/pkg/apis" \
  "kf:v1alpha1" \
  --go-header-file "${root_dir}/pkg/kf/internal/tools/option-builder/LICENSE_HEADER.go.txt"

# For some reason the fake doesn't have the right imports (it's missing
# k8s.io/client-go/rest)
os_friendly_sed 's|"k8s.io/apimachinery/pkg/runtime"|"k8s.io/apimachinery/pkg/runtime"\n"k8s.io/client-go/rest"|g' "$(go env GOPATH)/src/${ROOT_PACKAGE}/pkg/client/injection/client/fake/fake.go"
gofmt -s -w .
