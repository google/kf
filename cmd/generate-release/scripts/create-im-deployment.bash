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

set -eux

# XXX: In general, accepting so many flags is a bad design. However this
# script will only be called by automation.

project_id=$1
cluster=$2
zone=$3
node_count=$4
machine_type=$5
network=$6
release_channel=$7
repo_ref=$8
image_type="cos_containerd"

if
  [ -z "${project_id}" ] ||
    [ -z "${cluster}" ] ||
    [ -z "${zone}" ] ||
    [ -z "${node_count}" ] ||
    [ -z "${machine_type}" ] ||
    [ -z "${network}" ] ||
    [ -z "${release_channel}" ] ||
    [ -z "${repo_ref}" ]
then
  echo "usage: $0 [PROJECT_ID] [CLOUDSDK_CONTAINER_CLUSTER] [CLOUDSDK_COMPUTE_ZONE] [NODE_COUNT] [MACHINE_TYPE] [NETWORK] [RELEASE_CHANNEL] [REPO_REF]"
  exit 1
fi

REPO_URL="https://github.com/google/kf"
TERRAFORM_DIR="cmd/generate-release/scripts/"
DEPLOYMENT_ZONE="us-central1"
SERVICE_ACCOUNT="infra-manager-sa@${project_id}.iam.gserviceaccount.com"

if ! gcloud iam service-accounts describe "${SERVICE_ACCOUNT}" --project="${project_id}" > /dev/null 2>&1; then
    echo "Service Account ${SERVICE_ACCOUNT} does not exist. Run [ scripts/create-im-sa.bash ${project_id} ]"
    exit 1
fi

gcloud infra-manager deployments apply "projects/${project_id}/locations/${DEPLOYMENT_ZONE}/deployments/${cluster}" \
    --service-account="projects/${project_id}/serviceAccounts/${SERVICE_ACCOUNT}" \
    --git-source-repo="${REPO_URL}" \
    --git-source-directory=${TERRAFORM_DIR} \
    --git-source-ref=${repo_ref} \
    --input-values=project_id=${project_id},deployment_name=${cluster},zone=${zone},network=${network},initial_node_count=${node_count},machine_type=${machine_type},image_type=${image_type},release_channel=${release_channel}

echo "IM deployment created."
