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
image_type="cos_containerd"

if
  [ -z "${project_id}" ] ||
    [ -z "${cluster}" ] ||
    [ -z "${zone}" ] ||
    [ -z "${node_count}" ] ||
    [ -z "${machine_type}" ] ||
    [ -z "${network}" ] ||
    [ -z "${release_channel}" ]
then
  echo "usage: $0 [PROJECT_ID] [CLOUDSDK_CONTAINER_CLUSTER] [CLOUDSDK_COMPUTE_ZONE] [NODE_COUNT] [MACHINE_TYPE] [NETWORK] [RELEASE_CHANNEL]"
  exit 1
fi

function retry() {
    local retries=$1
    local wait=$2

    # Remove the first two arguments (retries and wait) to isolate the command
    shift 2 

    local count=0
    until "$@"; do
        exit_code=$?
        count=$((count + 1))
        if [ $count -ge $retries ]; then
            echo "Command \"$1\" failed after $retries attempts."
            return $exit_code
        fi
        echo "Command failed. Retrying in $wait seconds... (Attempt $count/$retries)"
        sleep $wait
    done
    return 0
}

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"

echo "Initializing Terraform..."
terraform -chdir="${script_dir}" init -upgrade

echo "Applying Terraform configuration..."
retry 15 60 terraform -chdir="${script_dir}" apply \
        -var="project_id=${project_id}" \
        -var="deployment_name=${cluster}" \
        -var="zone=${zone}" \
        -var="network=${network}" \
        -var="initial_node_count=${node_count}" \
        -var="machine_type=${machine_type}" \
        -var="image_type=${image_type}" \
        -var="release_channel=${release_channel}" \
        -auto-approve
echo "Terraform apply completed successfully."
