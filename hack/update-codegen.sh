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

HACK_DIR=$(realpath $(dirname "$(realpath $0)"))
CODEGEN_PKG=$(go env GOPATH)/src/k8s.io/code-generator
CODEGEN_PACKAGES=$(cat $HACK_DIR/codegen-packages.txt)
KF_PACKAGE="github.com/google/kf"
KF_PACKAGE_LOCATION="$(go env GOPATH)/src/$KF_PACKAGE"
KF_RESOURCE="kf:v1alpha1"
BUILD_RESOURCE="build:v1alpha1"
HEADER_FILE=${KF_PACKAGE_LOCATION}/pkg/kf/internal/tools/option-builder/LICENSE_HEADER.go.txt
KNATIVE_CODEGEN_PKG=$(go env GOPATH)/src/knative.dev/pkg

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

for PACKAGE in ${CODEGEN_PACKAGES}; do
  if [ ! -d $(go env GOPATH)/src/${PACKAGE} ]; then
    install_package $PACKAGE
  fi
done

if [ "$(realpath $KF_PACKAGE_LOCATION)" != "$(git rev-parse --show-toplevel)" ]; then
    echo "The generator scripts aren't go module compatible (yet)." 1>&2
    exit 1
fi

# Fix issues due to using old k8s.io/client-go
# The generator wants to use a codec that is only available in a version of
# k8s.io/apimachinery that we can't yet use.
os_friendly_sed () {
  echo "Applying $1 to $2"
  if [ ! -e "$2" ]; then
    echo file not found: $2 2>&1
    exit 1
  fi
  sed "$1" "$2" > "$2.new"
  mv "$2.new" "$2"
}

${CODEGEN_PKG}/generate-groups.sh all \
  "$KF_PACKAGE/pkg/client" \
  "$KF_PACKAGE/pkg/apis" \
  "$KF_RESOURCE" \
  --go-header-file="$HEADER_FILE" \
  "${GENERATOR_FLAGS}"
[ $? -ne 0 ] && echo Error running code-generator 1>&2 && exit 1

${CODEGEN_PKG}/generate-groups.sh "deepcopy,client,informer,lister" \
  "$KF_PACKAGE/pkg/client/build" \
  "github.com/knative/build/pkg/apis" \
  "$BUILD_RESOURCE" \
  --go-header-file="$HEADER_FILE" \
  "${GENERATOR_FLAGS}"
[ $? -ne 0 ] && echo Error running code-generator 1>&2 && exit 1

codegen_sed() {
  local CLIENT=$1
  local TYPE=$2
  local VERSION=$3
  local FILE=$4

  if [ "$CLIENT" = "kf" ]; then
    CLIENT=""
  fi

  os_friendly_sed 's/scheme.Codecs.WithoutConversion()/scheme.Codecs/g' "$KF_PACKAGE_LOCATION/pkg/client/$CLIENT/clientset/versioned/typed/$TYPE/$VERSION/$FILE.go"
  os_friendly_sed 's/pt, //g' "$KF_PACKAGE_LOCATION/pkg/client/$CLIENT/clientset/versioned/typed/$TYPE/$VERSION/fake/fake_${FILE}.go"
}

TYPES="build_client build buildtemplate clusterbuildtemplate"
for type in ${TYPES}; do
  codegen_sed build build v1alpha1 $type
done

TYPES="kf_client $(ls ${KF_PACKAGE_LOCATION}/pkg/apis/kf/v1alpha1/ | grep 'types.go' | sed 's/_types.go//')"
for type in ${TYPES}; do
  codegen_sed kf kf v1alpha1 $type
done

# Do Knative injection generation
${KNATIVE_CODEGEN_PKG}/hack/generate-knative.sh \
  "injection" \
  "github.com/google/kf/pkg/client" \
  "github.com/google/kf/pkg/apis" \
  "kf:v1alpha1" \
  --go-header-file $HEADER_FILE

${KNATIVE_CODEGEN_PKG}/hack/generate-knative.sh \
  "injection" \
  "github.com/google/kf/pkg/client/build" \
  "github.com/knative/build/pkg/apis" \
  "build:v1alpha1" \
  --go-header-file $HEADER_FILE

# For some reason the fake doesn't have the right imports (it's missing
# k8s.io/client-go/rest)

knative_injection_sed (){
  local CLIENT=$1

  if [ "$CLIENT" = "kf" ]; then
    CLIENT=""
  fi
  os_friendly_sed 's|"k8s.io/apimachinery/pkg/runtime"|"k8s.io/apimachinery/pkg/runtime"\n"k8s.io/client-go/rest"|g' "${KF_PACKAGE_LOCATION}/pkg/client/$CLIENT/injection/client/fake/fake.go"
}

knative_injection_sed kf
knative_injection_sed build

gofmt -s -w .
