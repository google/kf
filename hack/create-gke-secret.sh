#!/usr/bin/env bash

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

set -eu

cd "${0%/*}"/..

project=''
sa_name=''
key_json=''


print_usage() {
  echo "Usage: [-p GCP_PROJECT] [-a SERVICE ACCOUNT EMAIL] [-k JSON KEY]"
  echo
  echo "Examples:"
  echo "  # create a new service account and key"
  echo "  ${0}"
  echo "  # use an existing key (don't create SA)"
  echo "  ${0} -k '{json here}'"
  echo "  # use an existing SA, creating a new key"
  echo "  ${0} -p my-project -a my-sa@my-project.google.com"
}

while getopts 'p:a:k:' flag; do
  case "${flag}" in
    p) project="${OPTARG}" ;;
    a) sa_name="${OPTARG}" ;;
    k) key_json="${OPTARG}" ;;
    *) print_usage
       exit 1 ;;
  esac
done

# if the user provided a key, we don't need to create an SA
if [ "${key_json}" = "" ]; then
  if [ "${service_account}" = "" ]; then
    if [ "${project}" = "" ]; then
      echo "Autodetecting project, use -p to override"
      project=$(gcloud config get-value project)
    fi

    sa_name_prefix=create-gke-secret-${RANDOM}
    sa_name="${sa_name_prefix}@${project}.iam.gserviceaccount.com"

    echo "creating a GCP ServiceAccount (${sa_name}) for Project ${project}"
    current_context=$(kubectl config current-context)
    gcloud iam service-accounts create ${sa_name_prefix} \
      --project ${project} \
      --description "gcr.io admin for ${current_context}" \
      --display-name "${current_context}"

    echo "adding role roles/storage.admin binding to ServiceAccount"
    gcloud projects add-iam-policy-binding ${project} \
        --member "serviceAccount:${sa_name}" \
        --role "roles/storage.admin"
  fi

  temp_dir=$(mktemp -d)
  key_path=${temp_dir}/key.json

  echo "Creating service account key"
  gcloud iam service-accounts keys create \
    --iam-account ${sa_name} ${key_path}
  secret_name=kf-gcr-key-${RANDOM}
  key_json=$(cat ${key_path})
  rm -rf ${temp_dir}
fi

echo "creating K8s Secret"
secret_name=kf-gcr-key-${RANDOM}
kubectl -nkf create secret docker-registry ${secret_name} \
  --docker-username=_json_key \
  --docker-server https://gcr.io \
  --docker-password="${key_json}"

echo "creating K8s ConfigMap to point to Secret"
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: config-secrets
  namespace: kf
data:
  build.imagePushSecrets: "${secret_name}"
EOF
