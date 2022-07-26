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

# Authenticate
/builder/kf.bash &> /dev/null

project_id=$1
domain=$2
zone=$3
ar_repo=$4

if [ -z "${project_id}" ] || [ -z "${domain}" ] || [ -z "${zone}" ] || [ -z "${ar_repo}" ]; then
  echo "usage: $0 [PROJECT_ID] [DOMAIN] [ZONE] [AR_REPO]"
  exit 1
fi

# Convert zone into a region
region=$(echo "${zone}" | cut -d'-' -f1,2)

registry=${region}-docker.pkg.dev/${project_id}/${ar_repo}

echo "Domain: ${domain}"
echo "Container Registry: ${registry}"

# Patch the defaults
kubectl patch \
  configmaps config-defaults \
  -n kf \
  -p "{\"data\":{\"spaceContainerRegistry\":\"${registry}\",\"spaceClusterDomains\":\"- domain: ${domain}\"}}"
