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

export GO111MODULE=off

GENERATOR_FLAGS=""
while getopts "v" opt; do
  case $opt in
    v)
      set -x
      GENERATOR_FLAGS="-v 5 ${GENERATOR_FLAGS}"
      ;;
  esac
done

readonly CODEGEN_PKG=$(go env GOPATH)/src/k8s.io/code-generator
readonly KF_PACKAGE="github.com/google/kf"
readonly KF_PACKAGE_LOCATION="$(go env GOPATH)/src/$KF_PACKAGE"
readonly KF_RESOURCE="kf:v1alpha1"
readonly BUILD_RESOURCE="build:v1alpha1"
readonly HEADER_FILE=${KF_PACKAGE_LOCATION}/pkg/kf/internal/tools/option-builder/LICENSE_HEADER.go.txt

if [ ! -d "$KF_PACKAGE_LOCATION" ]; then
  echo "Cannot find go package $KF_PACKAGE" 1>&2
  exit 1
fi

install_package() {
  local PACKAGE=$1

  echo installing $PACKAGE

  pushd $(go env GOPATH) &>/dev/null
  go get -u $PACKAGE/...
  popd &>/dev/null
}

CODEGEN_PACKAGES=$(cat $(realpath $(dirname "$(realpath $0)")/codegen-packages.txt))
for PACKAGE in $CODEGEN_PACKAGES; do
  if [ ! -d $(go env GOPATH)/src/${PACKAGE} ]; then
    install_package $PACKAGE
  fi
done

if [ "$(realpath $KF_PACKAGE_LOCATION)" != "$(git rev-parse --show-toplevel)" ]; then
    echo "The generator scripts aren't go module compatible (yet)." 1>&2
    exit 1
fi


# run the code-generator entrypoint script
run_code_generator () {

  CLIENT_PACKAGE=$1
  API_PACKAGE=$2
  CUSTOM_RESOURCE=$3

  ${CODEGEN_PKG}/generate-groups.sh all \
    "$CLIENT_PACKAGE" \
    "$API_PACKAGE" \
    "$CUSTOM_RESOURCE" \
    --go-header-file="$HEADER_FILE" \
    "${GENERATOR_FLAGS}"

  ret=$?
  if [ $ret -ne 0 ]; then
    echo Error running code-generator 1>&2
    exit 1
  fi
}

# Fix issues due to using old k8s.io/client-go
# The generator wants to use a codec that is only available in a version of
# k8s.io/apimachinery that we can't yet use.
os_friendly_sed () {
  echo "Applying $1 to $2"
  sed "$1" "$2" > "$2.new"
  mv "$2.new" "$2"
}

run_code_generator "$KF_PACKAGE/pkg/client" "$KF_PACKAGE/pkg/apis" "$KF_RESOURCE"
#run_code_generator "$KF_PACKAGE/pkg/client" "github.com/knative/build/pkg/apis" "BUILD_RESOURCE"


${CODEGEN_PKG}/generate-groups.sh "deepcopy,client,informer,lister" \
  "github.com/google/kf/pkg/client/build" \
  "github.com/knative/build/pkg/apis" \
  "build:v1alpha1" \
  --go-header-file="$HEADER_FILE" \
  "${GENERATOR_FLAGS}"


TYPES="build_client build buildtemplate clusterbuildtemplate"
echo TYPES $TYPES
for type in ${TYPES}; do
  os_friendly_sed 's/scheme.Codecs.WithoutConversion()/scheme.Codecs/g' "${KF_PACKAGE_LOCATION}/pkg/client/build/clientset/versioned/typed/build/v1alpha1/${type}.go"
  os_friendly_sed 's/pt, //g' "$(go env GOPATH)/src/${KF_PACKAGE}/pkg/client/build/clientset/versioned/typed/build/v1alpha1/fake/fake_${type}.go"
done

TYPES="kf_client $(ls ${KF_PACKAGE_LOCATION}/pkg/apis/kf/v1alpha1/ | grep 'types.go' | sed 's/_types.go//')"
for type in ${TYPES}; do
  os_friendly_sed 's/scheme.Codecs.WithoutConversion()/scheme.Codecs/g' "${KF_PACKAGE_LOCATION}/pkg/client/clientset/versioned/typed/kf/v1alpha1/${type}.go"
  os_friendly_sed 's/pt, //g' "$(go env GOPATH)/src/${KF_PACKAGE}/pkg/client/clientset/versioned/typed/kf/v1alpha1/fake/fake_${type}.go"
done


# Do Knative injection generation
KNATIVE_CODEGEN_PKG=$(go env GOPATH)/src/knative.dev/pkg

${KNATIVE_CODEGEN_PKG}/hack/generate-knative.sh "injection" \
  "github.com/google/kf/pkg/client" "github.com/google/kf/pkg/apis" \
  "kf:v1alpha1" \
  --go-header-file "${KF_PACKAGE_LOCATION}/pkg/kf/internal/tools/option-builder/LICENSE_HEADER.go.txt"

${KNATIVE_CODEGEN_PKG}/hack/generate-knative.sh "injection" \
  "github.com/google/kf/pkg/client/build" "github.com/knative/build/pkg/apis" \
  "build:v1alpha1" \
  --go-header-file "${KF_PACKAGE_LOCATION}/pkg/kf/internal/tools/option-builder/LICENSE_HEADER.go.txt"

# For some reason the fake doesn't have the right imports (it's missing
# k8s.io/client-go/rest)
os_friendly_sed 's|"k8s.io/apimachinery/pkg/runtime"|"k8s.io/apimachinery/pkg/runtime"\n"k8s.io/client-go/rest"|g' "$(go env GOPATH)/src/${KF_PACKAGE}/pkg/client/injection/client/fake/fake.go"

os_friendly_sed 's|"k8s.io/apimachinery/pkg/runtime"|"k8s.io/apimachinery/pkg/runtime"\n"k8s.io/client-go/rest"|g' "$(go env GOPATH)/src/${KF_PACKAGE}/pkg/client/build/injection/client/fake/fake.go"

gofmt -s -w .
