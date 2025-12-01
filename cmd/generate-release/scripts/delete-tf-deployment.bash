#!/usr/bin/env bash

# This has been added to work around spurious problems with deleting GKE cluster
# at the end of integration tests (usualy due to "operation in progress")

set -eux

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

# retry will run the command, if it fails, it will sleep and then try again:
# args:
# 1: command: The command to run. (required)
# 2: retry_count: The number of times to attempt the command. (defaults to 3)
# 3: sleep_seconds: The number of seconds to sleep after a failure. (defaults to 1)
function retry() {
    command=${1:-x}
    retry_count=${2:-3}
    sleep_seconds=${3:-1}

    if [ "${command}" = "x" ]; then
        echo "command missing"
        exit 1
    fi

    n=0
    until [ "$n" -ge ${retry_count} ]
    do
       ${command} && break
       echo "failed..."
       n=$((n+1))
       sleep ${sleep_seconds}
    done

    if [ ${n} = ${retry_count} ];then
        echo "\"${command}\" failed too many times: (${retry_count})"
        exit 1
    fi
}

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"

retry "terraform apply -chdir="${script_dir}" \
        -var="project_id=${project_id}" \
        -var="deployment_name=${cluster}" \
        -var="zone=${zone}" \
        -var="network=${network}" \
        -var="initial_node_count=${node_count}" \
        -var="machine_type=${machine_type}" \
        -var="image_type=${image_type}" \
        -var="release_channel=${release_channel}" \
        -auto-approve" 15 60
    