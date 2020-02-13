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

cd "${0%/*}"/..

# https://github.com/kubernetes-sigs/kubebuilder/issues/359
# K8s code generation is broken when used in projects that use Go modules.
# Until that is fixed, this script will symlink the project into your GOPATH
# and remove the symlink when it is done. A GOPATH is required.
if [[ -z "$GOPATH" ]]; then
  echo "GOPATH must be set"
  exit 1
fi

# Symlink the project into the GOPATH.
# Required until https://github.com/kubernetes-sigs/kubebuilder/issues/359 is fixed
mkdir -p $GOPATH/src/github.com/google
olddir=$(pwd)
pushd $GOPATH/src/github.com/google

  if [[ "$olddir" == "$(pwd)/kf" ]]; then
    # already in gopath, no symlink needed
    echo
  elif [[ -L kf ]]; then
    if [[ "$(readlink kf)" != "$olddir" ]]; then
      echo link points somewhere else, updating
      rm kf
      ln -s $olddir kf
    fi
  else
    echo creating gopath symlink
    ln -s $olddir kf
  fi
popd

export GO111MODULE=on

GENERATOR_FLAGS=""
while getopts "v" opt; do
  case $opt in
    v)
      set -x
      GENERATOR_FLAGS="-v 5 ${GENERATOR_FLAGS}"
      shift
      ;;
  esac
done

KF_PACKAGE="github.com/google/kf"
KF_PACKAGE_LOCATION="./"
KF_RESOURCE="kf:v1alpha1"
TEKTONCD_PIPELINE_RESOURCE="pipeline:v1alpha1"
SERVING_RESOURCE="serving:v1alpha1"
HEADER_FILE=${KF_PACKAGE_LOCATION}/pkg/kf/internal/tools/option-builder/LICENSE_HEADER.go.txt

GENS=$1
if [ "$GENS" = "" ]; then
  GENS="all"
fi

download-scripts() {
  commit=$(cat go.mod | grep code-generator | grep =\> | tr '-' ' ' | awk '{print $NF}')
  echo downloading code-generator script at commit $commit

  CODEGEN_PKG=vendor/k8s.io/code-generator
  curl -sLOJ https://raw.githubusercontent.com/kubernetes/code-generator/${commit}/generate-groups.sh
  chmod +x generate-groups.sh
  mkdir -p $CODEGEN_PKG
  mv generate-groups.sh $CODEGEN_PKG/generate-groups.sh

  commit=$(cat go.mod | grep knative.dev/pkg | tr '-' ' ' | awk '{print $NF}')
  echo downloading knative-injection-generator script at commit $commit

  KNATIVE_CODEGEN_PKG=vendor/knative.dev/pkg/hack
  curl -sLOJ https://raw.githubusercontent.com/knative/pkg/${commit}/hack/generate-knative.sh
  chmod +x generate-knative.sh
  mkdir -p $KNATIVE_CODEGEN_PKG
  mv generate-knative.sh $KNATIVE_CODEGEN_PKG/generate-knative.sh
}

code-generator-gen() {

  GENERATORS=$1
  OUTPUT_PACKAGE=$2
  API_PACKAGE=$3
  GROUP_VERSION=$4

  echo "k8s code gen for $API_PACKAGE at $GROUP_VERSION -> $OUTPUT_PACKAGE"
  ${CODEGEN_PKG}/generate-groups.sh \
    "$GENERATORS" \
    "$OUTPUT_PACKAGE" \
    "$API_PACKAGE" \
    "$GROUP_VERSION" \
    --go-header-file="$HEADER_FILE" \
    ${GENERATOR_FLAGS}
  [ $? -ne 0 ] && echo Error running code-generator 1>&2 && exit 1

  return 0
}

knative-injection-gen() {

  GENERATORS=$1
  OUTPUT_PACKAGE=$2
  API_PACKAGE=$3
  GROUP_VERSION=$4

  echo "knative injection gen for $API_PACKAGE at $GROUP_VERSION -> $OUTPUT_PACKAGE"
  ${KNATIVE_CODEGEN_PKG}/generate-knative.sh \
    "$GENERATORS" \
    "$OUTPUT_PACKAGE" \
    "$API_PACKAGE" \
    "$GROUP_VERSION" \
    --go-header-file $HEADER_FILE
  [ $? -ne 0 ] && echo Error running injection-generator 1>&2 && exit 1

  return 0
}

kf-code-gen() {
  code-generator-gen \
    all \
    "$KF_PACKAGE/pkg/client" \
    "$KF_PACKAGE/pkg/apis" \
    "$KF_RESOURCE"
}

kf-knative-gen() {
  knative-injection-gen \
    "injection" \
    "github.com/google/kf/pkg/client" \
    "github.com/google/kf/pkg/apis" \
    "kf:v1alpha1"
}

tekton-pipeline-code-gen() {
  code-generator-gen \
    "deepcopy,client,informer,lister" \
    "github.com/google/kf/third_party/tektoncd-pipeline/pkg/client" \
    "github.com/google/kf/third_party/tektoncd-pipeline/pkg/apis" \
    "$TEKTONCD_PIPELINE_RESOURCE"
}

tekton-pipeline-knative-gen() {
  knative-injection-gen \
    "injection" \
    "github.com/google/kf/third_party/tektoncd-pipeline/pkg/client" \
    "github.com/google/kf/third_party/tektoncd-pipeline/pkg/apis" \
    "$TEKTONCD_PIPELINE_RESOURCE"
}

kserving-code-gen() {
  code-generator-gen \
    "deepcopy,client,informer,lister" \
    "github.com/google/kf/third_party/knative-serving/pkg/client" \
    "github.com/google/kf/third_party/knative-serving/pkg/apis" \
    "$SERVING_RESOURCE"
}

kserving-knative-gen() {
  knative-injection-gen \
    "injection" \
    "github.com/google/kf/third_party/knative-serving/pkg/client" \
    "github.com/google/kf/third_party/knative-serving/pkg/apis" \
    "$SERVING_RESOURCE"
}

svccat-codegen() {
  code-generator-gen \
    "deepcopy,client,informer,lister" \
    "$KF_PACKAGE/pkg/client/servicecatalog" \
    "github.com/poy/service-catalog/pkg/apis" \
    "servicecatalog:v1beta1"
}

svccat-knative-gen() {
  knative-injection-gen \
    "injection" \
    "github.com/google/kf/pkg/client/servicecatalog" \
    "github.com/poy/service-catalog/pkg/apis" \
    "servicecatalog:v1beta1"
}

REPO_ROOT="${REPO_ROOT:-$(git rev-parse --show-toplevel)}"
if [[ "$REPO_ROOT" = "" ]]; then
  echo could not find REPO_ROOT 1>&2
  exit 1
fi
cd "${REPO_ROOT}"

# enable modules and the proxy cache
export GO111MODULE="on"
GOPROXY="${GOPROXY:-https://proxy.golang.org}"
export GOPROXY

go mod vendor
export GO111MODULE="off"

# fake being in a gopath
FAKE_GOPATH="$(mktemp -d)"
trap 'rm -rf ${FAKE_GOPATH}' EXIT

FAKE_REPOPATH="${FAKE_GOPATH}/src/github.com/google/kf"
mkdir -p "$(dirname "${FAKE_REPOPATH}")" && ln -s "${REPO_ROOT}" "${FAKE_REPOPATH}"
export GOPATH="${FAKE_GOPATH}"
cd "${FAKE_REPOPATH}"

download-scripts

case $GENS in
  k8s)
    kf-code-gen
    tekton-pipeline-code-gen
    svccat-code-gen
    ;;
  knative)
    kf-knative-gen
    tekton-pipeline-knative-gen
    kserving-knative-gen
    svccat-knative-gen
    ;;
  kf)
    kf-code-gen
    kf-knative-gen
    ;;
  tekton-pipeline)
    tekton-pipeline-code-gen
    tekton-pipeline-knative-gen
    ;;
  kserving)
    kserving-code-gen
    kserving-knative-gen
    ;;
  svccat)
    svccat-codegen
    svccat-knative-gen
    ;;
  all)
    rm -fr "${REPO_ROOT}/pkg/client"
    kf-code-gen
    kf-knative-gen
    tekton-pipeline-code-gen
    tekton-pipeline-knative-gen
    kserving-code-gen
    kserving-knative-gen
    svccat-codegen
    svccat-knative-gen
    ;;
  *)
    echo "invalid codegen argument: $GENS"
    exit 1
esac

gofmt -s -w .
[ $? -ne 0 ] && echo Error running gofmt 1>&2 && exit 1

exit 0
