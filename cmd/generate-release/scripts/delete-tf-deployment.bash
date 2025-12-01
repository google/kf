#!/usr/bin/env bash

# This has been added to work around spurious problems with deleting GKE cluster
# at the end of integration tests (usually due to "operation in progress")

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

# Usage: retry <retries> <wait_seconds> <command> [args...]
function retry() {
    local retries=$1
    local wait=$2
    shift 2 # Remove the first two arguments (retries and wait) to isolate the command

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

echo "Destroying Terraform configuration..."
retry 15 60 terraform -chdir="${script_dir}" destroy \
        -var="project_id=${project_id}" \
        -var="deployment_name=${cluster}" \
        -var="zone=${zone}" \
        -var="network=${network}" \
        -var="initial_node_count=${node_count}" \
        -var="machine_type=${machine_type}" \
        -var="image_type=${image_type}" \
        -var="release_channel=${release_channel}" \
        -auto-approve