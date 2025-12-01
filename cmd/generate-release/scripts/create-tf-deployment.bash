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

set -ex

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

# Create/Update the deployment
retry_count=3
n=0
until [ "$n" -ge ${retry_count} ]; do
    terraform init -upgrade && \
    script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
    terraform init -upgrade -chdir="${script_dir}" && \
    terraform apply -chdir="${script_dir}" \
      -var="project_id=${project_id}" \
      -var="deployment_name=${cluster}" \
      -var="zone=${zone}" \
      -var="network=${network}" \
      -var="initial_node_count=${node_count}" \
      -var="machine_type=${machine_type}" \
      -var="image_type=${image_type}" \
      -var="release_channel=${release_channel}" \
      -auto-approve && \
      break

  n=$((n + 1))
  echo -e "import random\n\nimport time\ntime.sleep(random.randint(30,90))" | python3
done

if [ ${n} = ${retry_count} ]; then
  echo "create deployment failed too many times (${retry_count})"
  exit 1
fi
