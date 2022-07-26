#!/usr/bin/env bash
# Copyright 2022 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


# This script can be ran without any arguments. Here are the following
# optional inputs (all envs).
# PROJECT_ID - defaults to current project ID.
# SERVICE_ACCOUNT_NAME - Defaults to random name with kf-gcp-broker prefix

set -eux

export PROJECT_ID=${PROJECT_ID:-$(gcloud config get-value project)}
export SERVICE_ACCOUNT_NAME=${SERVICE_ACCOUNT_NAME:-kf-gcp-broker-${RANDOM}}

temp_dir=$(mktemp -d)
finish() {
    rm -rf ${temp_dir}
}
trap finish EXIT

# alter_service_account.py alters the YAML for the helm chart. It inserts the
# JSON key into the broker.service_account_json value.
cat << EOF >> ${temp_dir}/alter_service_account.py
import yaml
import sys

if len(sys.argv) != 2:
    print("must pass path to key file")
    sys.exit(1)
key_path  =sys.argv[1]

with open("values.yaml", 'r') as f:
    values = yaml.safe_load(f)


with open(key_path, 'r') as f:
    values["broker"]["service_account_json"] = f.read()

with open("values.yaml", 'w') as f:
    yaml.dump(values, f)
EOF

gcloud iam service-accounts create ${SERVICE_ACCOUNT_NAME}
gcloud iam service-accounts keys create ${temp_dir}/key.json --iam-account ${SERVICE_ACCOUNT_NAME}@${PROJECT_ID}.iam.gserviceaccount.com
gcloud projects add-iam-policy-binding ${PROJECT_ID} --member serviceAccount:${SERVICE_ACCOUNT_NAME}@${PROJECT_ID}.iam.gserviceaccount.com --role "roles/owner"
gcloud services enable cloudresourcemanager.googleapis.com iam.googleapis.com --project ${PROJECT_ID}

pushd ${temp_dir}
    git clone --depth=1 "https://github.com/GoogleCloudPlatform/gcp-service-broker"
    pushd gcp-service-broker/deployments/helm/gcp-service-broker
        helm dependency update
        python3 ${temp_dir}/alter_service_account.py ${temp_dir}/key.json

        kubectl create namespace gcp-service-broker
        helm install gcp-service-broker --set svccat.register=false --namespace gcp-service-broker .

        # Wait a moment for things to be ready.
        sleep 60

        kf create-service-broker gcp-service-broker \
          "$(kubectl get secret gcp-service-broker-auth -n gcp-service-broker -o jsonpath='{.data.username}' | base64 --decode)" \
          "$(kubectl get secret gcp-service-broker-auth -n gcp-service-broker -o jsonpath='{.data.password}' | base64 --decode)" \
          "http://gcp-service-broker.gcp-service-broker.svc.cluster.local"
    popd
popd
