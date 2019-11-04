#!/usr/bin/env sh

# Copyright 2019 Google LLC
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
cd "${0%/*}"/..

project=$(gcloud config get-value project)
sa_name_prefix=create-gke-secret-${RANDOM}
sa_name="${sa_name_prefix}@${project}.iam.gserviceaccount.com"

echo "creating a GCP ServiceAccount (${sa_name}) for Project ${project}"
gcloud iam service-accounts create ${sa_name_prefix} --project ${project}

echo "adding role roles/storage.admin binding to ServiceAccount"
gcloud projects add-iam-policy-binding ${project} \
    --member "serviceAccount:${sa_name}" \
    --role "roles/storage.admin"

echo "downloading key... (this will be deleted after)"
temp_dir=$(mktemp -d)
delete_temp() {
    rm -rf ${temp_dir}
}
trap delete_temp EXIT

key_path=${temp_dir}/key.json
gcloud iam service-accounts keys create \
  --iam-account ${sa_name} ${key_path}

echo "creating K8s Secret"
secret_name=kf-gcr-key-${RANDOM}
kubectl -nkf create secret docker-registry ${secret_name} \
  --docker-username=_json_key \
  --docker-server https://gcr.io \
  --docker-password="$(cat ${key_path})"

echo "creating K8s ConfigMap to point to Secret"
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: config-secrets
  namespace: kf
data:
  build.imagePushSecret: "${secret_name}"
EOF
