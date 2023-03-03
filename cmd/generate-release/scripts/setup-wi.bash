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

if [ -z ${1:+x} ] || [ -z ${2:+x} ]; then
  echo "usage: $0 [PROJECT_ID] [CLUSTER_NAME]"
  exit 1
fi

project_id=$1
cluster_name=$2

if [ ! "$(kubectl -nkf get configmap config-secrets -o=jsonpath='{.data}')" ]; then
  if [ ! "$(yes | gcloud beta iam service-accounts list --filter "name:${cluster_name}-sa")" ]; then
    echo "creating new GCP service-account"
    yes | gcloud beta iam service-accounts create \
            "${cluster_name}-sa" \
            --description "gcr.io admin for ${cluster_name}" \
            --display-name "${cluster_name}"

  else
    echo "using existing GCP service-account"
  fi

  if [ ! "$(gcloud projects get-iam-policy "${project_id}" --format json --filter "bindings.members:(serviceAccount:${cluster_name}-sa@${project_id}.iam.gserviceaccount.com) AND bindings.role:(roles/storage.admin)")" ]; then
      # Give service account role to read/write from GCR
      gcloud projects add-iam-policy-binding "${project_id}" \
        --member "serviceAccount:${cluster_name}-sa@${project_id}.iam.gserviceaccount.com" \
        --role "roles/storage.admin" \
        --format=none
  else
      echo "roles/storage.admin already present"
  fi

  if [ ! "$(gcloud projects get-iam-policy "${project_id}" --format json --filter "bindings.members:(serviceAccount:${cluster_name}-sa@${project_id}.iam.gserviceaccount.com) AND bindings.role:(roles/iam.serviceAccountAdmin)")" ]; then
  # Give service account role to access IAM
  gcloud projects add-iam-policy-binding "${project_id}" \
    --member "serviceAccount:${cluster_name}-sa@${project_id}.iam.gserviceaccount.com" \
    --role "roles/iam.serviceAccountAdmin" \
    --format=none
  else
      echo "roles/iam.serviceAccountAdmin already present"
  fi

  # Workload Identity
  # Link GSA with Kf KSA
  # https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity
  # NOTE: This assumes the following:
  # * Kf controller is in 'kf' namespace
  # * Kf controller has the KSA 'controller'
  gcloud iam service-accounts add-iam-policy-binding \
    --role roles/iam.workloadIdentityUser \
    --member "serviceAccount:${project_id}.svc.id.goog[kf/controller]" \
    --member "serviceAccount:${project_id}.svc.id.goog[cnrm-system/cnrm-controller-manager]" \
    "${cluster_name}-sa@${project_id}.iam.gserviceaccount.com"

  kubectl annotate serviceaccount \
    --namespace kf \
    --overwrite \
    controller \
    "iam.gke.io/gcp-service-account=${cluster_name}-sa@${project_id}.iam.gserviceaccount.com"

  echo "{\"apiVersion\":\"v1\",\"kind\":\"ConfigMap\",\"metadata\":{\"name\":\"config-secrets\", \"namespace\":\"kf\"},\"data\":{\"wi.googleServiceAccount\":\"${cluster_name}-sa@${project_id}.iam.gserviceaccount.com\"}}" | kubectl apply -f -

  # Annotate the Kf namespace with the KCC annotation.
  kubectl annotate --overwrite=true namespace kf cnrm.cloud.google.com/project-id=${project_id}
else
  echo "Config Map config-secrets already configured"
fi
