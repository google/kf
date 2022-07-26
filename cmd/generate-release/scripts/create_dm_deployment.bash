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

# Ensure the DM service account is a project owner. More information can
# be found here:
# https://cloud.google.com/deployment-manager/docs/configuration/set-access-control-resources#granting_permission_to_set_iam_policies
python3 /builder/setup_dm_service_account.py "${project_id}"

# Create/Update the deployment
retry_count=3
n=0
until [ "$n" -ge ${retry_count} ]; do
  # Determine the command: update or create
  if [ ! "$(gcloud deployment-manager deployments list --filter "name<=${cluster} AND name>=${cluster}")" ]; then
    echo "deployment ${cluster} does not exists, will create deployment"
    command="create"
  else
    echo "deployment ${cluster} already exists, will update existing deployment"
    command="update"
  fi

  gcloud deployment-manager deployments "${command}" \
    "${cluster}" \
    --properties "zone:${zone},initialNodeCount:${node_count},machineType:${machine_type},network:${network},releaseChannel:${release_channel}" \
    --template /kf/bin/deployment-manager/cluster.py &&
    break
  n=$((n + 1))
  sleep 1
done

if [ ${n} = ${retry_count} ]; then
  echo "create deployment failed too many times (${retry_count})"
  exit 1
fi
