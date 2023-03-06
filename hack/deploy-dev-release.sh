#!/usr/bin/env bash

# Copyright 2020 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the License);
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an AS IS BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -eu
cd "${0%/*}"/..

# VERSION is required for create-dev-release.sh. If it's empty, then just use
# the logic in create-dev-release.sh.
# Do the same for EXTRA_TAG
export VERSION=${VERSION:=""}
export EXTRA_TAG=${EXTRA_TAG:=""}
release=$(./hack/create-dev-release.sh)

cluster_name="${CLUSTER_NAME:=kf-${RANDOM}-${USER}}"
zone="${GCP_COMPUTE_ZONE:=us-central1-a}"
machine="${MACHINE_TYPE:=e2-standard-4}"
asm_managed="${ASM_MANAGED:=true}"

git submodule update --init --recursive

gcloud builds submit \
  --no-source \
  --config=<(gsutil cat "gs://${release}/cloud-build/fresh-cluster.yaml") \
  --substitutions="_CLOUDSDK_COMPUTE_ZONE=${zone},_CLOUDSDK_CONTAINER_CLUSTER=${cluster_name},_MACHINE_TYPE=${machine},_ASM_MANAGED=${asm_managed}"

gcloud container clusters get-credentials "${cluster_name}" --zone "${zone}"
