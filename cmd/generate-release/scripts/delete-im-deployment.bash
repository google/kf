#!/usr/bin/env bash

set -eux

project_id=$1
cluster=$2
zone=$3

if
  [ -z "${project_id}" ] ||
    [ -z "${cluster}" ] ||
    [ -z "${zone}" ]
then
  echo "usage: $0 [PROJECT_ID] [CLOUDSDK_CONTAINER_CLUSTER] [CLOUDSDK_COMPUTE_ZONE]"
  exit 1
fi

DEPLOYMENT_ZONE="us-central1"

gcloud infra-manager deployments delete \
    projects/${project_id}/locations/${DEPLOYMENT_ZONE}/deployments/${cluster} \
    --quiet
echo "IM deployment deleted."
