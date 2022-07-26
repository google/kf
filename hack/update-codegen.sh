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

set -eux

cd "${0%/*}"/..
KF_PACKAGE="github.com/google/kf/v2"
HEADER_FILE="./pkg/kf/internal/tools/option-builder/LICENSE_HEADER.go.txt"

# https://github.com/kubernetes-sigs/kubebuilder/issues/359
# K8s code generation is broken when used in projects that use Go modules.
# Until that is fixed, this script will symlink the project into your GOPATH
# and remove the symlink when it is done. A GOPATH is required.

# Fetch dependencies
GO111MODULE="on" GOPROXY="${GOPROXY:-https://proxy.golang.org}" go mod vendor

# Set up fake GOPATH
FAKE_GOPATH="$(mktemp -d)"
mkdir -p ${FAKE_GOPATH}/src/
cp -r ./vendor/* ${FAKE_GOPATH}/src/
FAKE_REPOPATH="${FAKE_GOPATH}/src/${KF_PACKAGE}"
mkdir -p "$(dirname "${FAKE_REPOPATH}")" && ln -s "$PWD" "${FAKE_REPOPATH}"

# Cleanup on exit
trap 'rm -rf ${FAKE_GOPATH}' EXIT

# gopath-run sets up environment variables to run without modules.
gopath-run() {
  GO111MODULE="off" GOPATH="${FAKE_GOPATH}" go run $@
}

# Taken from https://github.com/kubernetes/code-generator/blob/master/generate-groups.sh.
function codegen::join() { local IFS="$1"; shift; echo "$*"; }

# Adapted from https://github.com/kubernetes/code-generator/blob/master/generate-groups.sh
# with changes to support Knative injection generation and go modules.
code-generator-gen() {
  GENS="$1"
  OUTPUT_PKG="$2"
  APIS_PKG="$3"
  GROUPS_WITH_VERSIONS="$4"

  echo "Kubernetes code gen for $APIS_PKG at $GROUPS_WITH_VERSIONS -> $OUTPUT_PKG"

  # enumerate group versions
  FQ_APIS=() # e.g. k8s.io/api/apps/v1
  for GVs in ${GROUPS_WITH_VERSIONS}; do
    IFS=: read G Vs <<<"${GVs}"

    # enumerate versions
    for V in ${Vs//,/ }; do
      FQ_APIS+=(${APIS_PKG}/${G}/${V})
    done
  done

  INPUT_DIRS="$(codegen::join , "${FQ_APIS[@]}")"

  GENCMD="./vendor/k8s.io/code-generator/cmd"
  if [ "${GENS}" = "all" ] || grep -qw "deepcopy" <<<"${GENS}"; then
    echo "Generating deepcopy funcs"
    gopath-run "${GENCMD}/deepcopy-gen" \
      --input-dirs "${INPUT_DIRS}" \
      -O zz_generated.deepcopy \
      --bounding-dirs "${APIS_PKG}" \
      --go-header-file="${HEADER_FILE}"
  fi

  if [ "${GENS}" = "all" ] || grep -qw "client" <<<"${GENS}"; then
    echo "Generating clientset for ${GROUPS_WITH_VERSIONS} at ${OUTPUT_PKG}/clientset"
    gopath-run "${GENCMD}/client-gen" \
      --clientset-name "versioned" \
      --input-base="" \
      --input "${INPUT_DIRS}" \
      --output-package "${OUTPUT_PKG}/clientset" \
      --go-header-file="${HEADER_FILE}"
  fi

  if [ "${GENS}" = "all" ] || grep -qw "lister" <<<"${GENS}"; then
    echo "Generating listers for ${GROUPS_WITH_VERSIONS} at ${OUTPUT_PKG}/listers"
    gopath-run "${GENCMD}/lister-gen" \
      --input-dirs "${INPUT_DIRS}" \
      --output-package "${OUTPUT_PKG}/listers" \
      --go-header-file="${HEADER_FILE}"
  fi

  if [ "${GENS}" = "all" ] || grep -qw "informer" <<<"${GENS}"; then
    echo "Generating informers for ${GROUPS_WITH_VERSIONS} at ${OUTPUT_PKG}/informers"
    gopath-run "${GENCMD}/informer-gen" \
     --input-dirs "${INPUT_DIRS}" \
     --versioned-clientset-package "${OUTPUT_PKG}/clientset/versioned" \
     --listers-package "${OUTPUT_PKG}/listers" \
     --output-package "${OUTPUT_PKG}/informers" \
     --go-header-file="${HEADER_FILE}"
  fi

  if [ "${GENS}" = "all" ] || grep -qw "injection" <<<"${GENS}"; then
    echo "Generating injection for ${GROUPS_WITH_VERSIONS} at ${OUTPUT_PKG}/injection"
    gopath-run ./vendor/knative.dev/pkg/codegen/cmd/injection-gen \
      --input-dirs "${INPUT_DIRS}" \
      --versioned-clientset-package "${OUTPUT_PKG}/clientset/versioned" \
      --external-versions-informers-package "${OUTPUT_PKG}/informers/externalversions" \
      --listers-package "${OUTPUT_PKG}/listers" \
      --output-package "${OUTPUT_PKG}/injection" \
      --go-header-file="${HEADER_FILE}"
  fi
}

# Do generation for all clients.
rm -fr "./pkg/client"

pushd "${FAKE_REPOPATH}"
  code-generator-gen \
    all \
    "$KF_PACKAGE/pkg/client/kf" \
    "$KF_PACKAGE/pkg/apis" \
    "kf:v1alpha1"

  code-generator-gen \
    "deepcopy" \
    "$KF_PACKAGE/pkg/client/kf" \
    "$KF_PACKAGE/pkg/apis" \
    "kf:config"

  code-generator-gen \
    all \
    "$KF_PACKAGE/pkg/client/networking" \
    "$KF_PACKAGE/pkg/apis" \
    "networking:v1alpha3"

  code-generator-gen \
    all \
    "$KF_PACKAGE/pkg/client/kube" \
    "k8s.io/api" \
    "networking:v1"

  code-generator-gen \
    all \
    "$KF_PACKAGE/pkg/client/kube-aggregator" \
    "k8s.io/kube-aggregator/pkg/apis" \
    "apiregistration:v1"
popd

# Cleanup
./hack/tidy.sh
