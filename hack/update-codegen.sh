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

set -e

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

HACK_DIR="${0%/*}"
CODEGEN_PKG=vendor/k8s.io/code-generator

KF_PACKAGE="github.com/google/kf"
KF_PACKAGE_LOCATION="./"
KF_RESOURCE="kf:v1alpha1"
BUILD_RESOURCE="build:v1alpha1"
HEADER_FILE=${KF_PACKAGE_LOCATION}/pkg/kf/internal/tools/option-builder/LICENSE_HEADER.go.txt
KNATIVE_CODEGEN_PKG=vendor/knative.dev/pkg

GENS=$1
if [ "$GENS" = "" ]; then
  GENS="all"
fi

code-generator-gen() {
  echo running code-generator
  commit=$(cat go.mod | grep code-generator | grep =\> | tr '-' ' ' | awk '{print $NF}')
  curl -LOJ https://raw.githubusercontent.com/kubernetes/code-generator/${commit}/generate-groups.sh
  chmod +x generate-groups.sh
  mv generate-groups.sh ./vendor/k8s.io/code-generator

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
}

knative-injection-gen() {
  echo running knative-injection-generator
  commit=$(cat go.mod | grep knative.dev/pkg | tr '-' ' ' | awk '{print $NF}')
  curl -LOJ https://raw.githubusercontent.com/knative/pkg/${commit}/hack/generate-knative.sh
  chmod +x generate-knative.sh
  mkdir -p vendor/knative.dev/pkg/hack
  mv generate-knative.sh ./vendor/knative.dev/pkg/hack
  # Do Knative injection generation
  bash -x ${KNATIVE_CODEGEN_PKG}/hack/generate-knative.sh \
    "injection" \
    "github.com/google/kf/pkg/client" \
    "github.com/google/kf/pkg/apis" \
    "kf:v1alpha1" \
    --go-header-file $HEADER_FILE

  bash -x ${KNATIVE_CODEGEN_PKG}/hack/generate-knative.sh \
    "injection" \
    "github.com/google/kf/pkg/client/build" \
    "github.com/knative/build/pkg/apis" \
    "build:v1alpha1" \
    --go-header-file $HEADER_FILE
}

GO111MODULE=on
go mod vendor

case $GENS in
  k8s)
    code-generator-gen
    ;;
  knative)
    knative-injection-gen
    ;;
  all)
    code-generator-gen
    knative-injection-gen
    ;;
esac




gofmt -s -w .
