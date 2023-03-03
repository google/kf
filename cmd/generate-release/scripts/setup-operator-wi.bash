#!/usr/bin/env bash

set -eux

if [ -z ${1:+x} ] || [ -z ${2:+x} ]; then
  echo "usage: $0 [PROJECT_ID] [CLUSTER_NAME]"
  exit 1
fi

project_id=$1
cluster_name=$2

if [ ! "$(yes | gcloud beta iam service-accounts list --filter "name:${cluster_name}-sa")" ]; then
  echo "creating new GCP service-account"
  yes | gcloud beta iam service-accounts create \
          "${cluster_name}-sa" \
          --description "gcr.io admin for ${cluster_name}" \
          --display-name "${cluster_name}"

else
  echo "using existing GCP service-account"
fi

# Give service account role to read/write from GCR
gcloud projects add-iam-policy-binding "${project_id}" \
  --member "serviceAccount:${cluster_name}-sa@${project_id}.iam.gserviceaccount.com" \
  --role "roles/storage.admin" \
  --format=none

# Give service account role to access IAM
gcloud projects add-iam-policy-binding "${project_id}" \
  --member "serviceAccount:${cluster_name}-sa@${project_id}.iam.gserviceaccount.com" \
  --role "roles/iam.serviceAccountAdmin" \
  --format=none

# Workload Identity
# Link GSA with Kf KSA
# https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity
# NOTE: This assumes the following:
# * Kf controller is in 'kf' namespace
# * Kf controller has the KSA 'controller'
# * KCC controller is in cnrm-system namespace
# * KCC controller has the KSA 'cnrm-controller-manager'
gcloud iam service-accounts add-iam-policy-binding \
  --role roles/iam.workloadIdentityUser \
  --member "serviceAccount:${project_id}.svc.id.goog[kf/controller]" \
  "${cluster_name}-sa@${project_id}.iam.gserviceaccount.com"

gcloud iam service-accounts add-iam-policy-binding \
  --role roles/iam.workloadIdentityUser \
  --member "serviceAccount:${project_id}.svc.id.goog[cnrm-system/cnrm-controller-manager]" \
  "${cluster_name}-sa@${project_id}.iam.gserviceaccount.com"