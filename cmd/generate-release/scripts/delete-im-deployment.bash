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

REPO_URL="https://github.com/google/kf"
REPO_BRANCH="jakweg/migrate-dm-to-tf" # TODO: change to main before merged
TERRAFORM_DIR="cmd/generate-release/scripts/"
DEPLOYMENT_ZONE="us-central1"
SERVICE_ACCOUNT="infra-manager-sa@${project_id}.iam.gserviceaccount.com"

echo "Deleting IM deployment..."
gcloud infra-manager deployments delete \
    projects/${project_id}/locations/${DEPLOYMENT_ZONE}/deployments/${cluster} \
    --quiet
echo "IM deployment deleted."
