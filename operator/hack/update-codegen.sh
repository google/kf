#!/usr/bin/env bash

# Copyright 2019 The Knative Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -euo pipefail

OPERATOR_PACKAGE=kf-operator/


# Set up fake GOPATH
FAKE_GOPATH="$(mktemp -d)"
mkdir -p ${FAKE_GOPATH}/src/
cp -r ./vendor/* ${FAKE_GOPATH}/src/
FAKE_REPOPATH="${FAKE_GOPATH}/src/${OPERATOR_PACKAGE}"
mkdir -p "$(dirname "${FAKE_REPOPATH}")" && ln -s "$PWD" "${FAKE_REPOPATH}"

# Cleanup on exit
trap 'rm -rf ${FAKE_GOPATH}' EXIT

GOPATH=${FAKE_GOPATH}

export GO111MODULE=on

source "$(dirname "${BASH_SOURCE[0]}")/../vendor/knative.dev/hack/library.sh"
readonly OPERATOR_REPO_DIR="${REPO_ROOT_DIR}/operator"
source "${OPERATOR_REPO_DIR}/hack/library.sh"

CODEGEN_PKG=${CODEGEN_PKG:-$(cd ${OPERATOR_REPO_DIR}; ls -d -1 ./vendor/k8s.io/code-generator 2>/dev/null || echo ../code-generator)}

KNATIVE_CODEGEN_PKG=${KNATIVE_CODEGEN_PKG:-$(cd ${OPERATOR_REPO_DIR}; ls -d -1 ./vendor/knative.dev/pkg 2>/dev/null || echo ../pkg)}

# generate the code with:
# --output-base    because this script should also be able to run inside the vendor dir of
#                  k8s.io/kubernetes. The output-base is needed for the generators to output into the vendor dir
#                  instead of the $GOPATH directly. For normal projects this can be dropped.
bash ${CODEGEN_PKG}/generate-groups.sh "deepcopy,client,informer,lister" \
  kf-operator/pkg/client kf-operator/pkg/apis \
  "operand:v1alpha1 kfsystem:v1alpha1" \
  --go-header-file ${OPERATOR_REPO_DIR}/hack/boilerplate/boilerplate.go.txt

bash ${KNATIVE_CODEGEN_PKG}/hack/generate-knative.sh "injection" \
  kf-operator/pkg/client kf-operator/pkg/apis \
  "operand:v1alpha1 kfsystem:v1alpha1" \
  --go-header-file ${OPERATOR_REPO_DIR}/hack/boilerplate/boilerplate.go.txt

# Depends on generate-groups.sh to install bin/deepcopy-gen
${GOPATH}/bin/deepcopy-gen \
  -O zz_generated.deepcopy \
  --go-header-file ${OPERATOR_REPO_DIR}/hack/boilerplate/boilerplate.go.txt \
  -i kf-operator/pkg/apis/kfsystem/kf \
  -i kf-operator/pkg/apis/kfsystem/v1alpha1 \

# Generate the mock interfaces
echo "Generating mock interfaces"
pushd ${OPERATOR_REPO_DIR}/pkg/operand
go run github.com/golang/mock/mockgen -source=api.go \
  -destination=mock/mock_api.go \
  -package=mock
popd

# Make sure our dependencies are up-to-date
${OPERATOR_REPO_DIR}/hack/update-deps.sh
