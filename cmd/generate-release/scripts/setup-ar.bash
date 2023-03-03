#!/usr/bin/env bash

set -ex

project_id=$1
repo_name=$2
zone=$3
gsa=$4

if \
  [ -z "${project_id}" ] \
  || [ -z "${repo_name}" ] \
  || [ -z "${zone}" ] \
  || [ -z "${gsa}" ]; then
  echo "usage: $0 [PROJECT_ID] [REPO_NAME] [ZONE] [GSA]"
  exit 1
fi

# Convert zone into a region
region=$(echo "${zone}" | cut -d'-' -f1,2)

# Ensure the Artifact Registry repo doesn't already exist.
if [ ! "$(gcloud artifacts repositories describe ${repo_name} --location ${region})" ]; then
    echo "Creating Artifact Registry repo ${repo_name} in ${region}..."
    gcloud artifacts repositories create \
      ${repo_name} \
      --repository-format=docker \
      --location=${region}
else
    echo "Artifact Registry repo ${repo_name} in ${region} already exists.  Moving on..."
fi

# Granting GSA permissions on AR repo
gcloud projects add-iam-policy-binding "${project_id}" \
  --member "serviceAccount:${gsa}" \
  --role "roles/artifactregistry.writer" \
  --format=none