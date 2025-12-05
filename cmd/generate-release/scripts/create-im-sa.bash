#!/usr/bin/env bash

# Copyright 2025 Google LLC
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

set -euo pipefail

if [ -z "$1" ]; then
    echo "Usage: $0 <PROJECT_ID>"
    exit 1
fi

PROJECT_ID=$1
SA_NAME="infra-manager-sa"
SA_EMAIL="${SA_NAME}@${PROJECT_ID}.iam.gserviceaccount.com"
DISPLAY_NAME="Infra Manager Provisioning Account"
DESCRIPTION="This Service Account is used by cloud build trigger kf-integ-tests-daily-rapid-head"

ROLES=(
    "roles/config.agent"                    # Cloud Infrastructure Manager Agent
    "roles/compute.networkAdmin"            # Compute Network Admin
    "roles/container.admin"                 # Kubernetes Engine Admin
    "roles/logging.logWriter"               # Logs Writer
    "roles/resourcemanager.projectIamAdmin" # Project IAM Admin
    "roles/iam.serviceAccountAdmin"         # Service Account Admin
    "roles/iam.serviceAccountUser"          # Service Account User
    "roles/storage.objectAdmin"             # Storage Object Admin
    "roles/storage.objectViewer"            # Storage Object Viewer
)

echo "Checking if Service Account ${SA_EMAIL} exists..."

if ! gcloud iam service-accounts describe "${SA_EMAIL}" --project="${PROJECT_ID}" > /dev/null 2>&1; then
    echo "Creating Service Account..."
    gcloud iam service-accounts create "${SA_NAME}" \
        --description="${DESCRIPTION}" \
        --display-name="${DISPLAY_NAME}" \
        --project="${PROJECT_ID}"
    echo "Service Account created."
else
    echo "Service Account already exists."
fi

echo "Assigning roles to ${SA_EMAIL}..."

for role in "${ROLES[@]}"; do
    echo "Assigning role: $role"
    # Use '|| true' to suppress errors if the binding already exists
    gcloud projects add-iam-policy-binding "${PROJECT_ID}" \
        --member="serviceAccount:${SA_EMAIL}" \
        --role="${role}" \
        --condition=None \
        --quiet > /dev/null
done

echo "Success! Service Account ${SA_EMAIL} is ready."