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

set -euxo pipefail

cd "${0%/*}"/../..
echo $(pwd)

_GCP_PROJECT_ID="${GCP_PROJECT_ID:}"
_RELEASE_BUCKET="${RELEASE_BUCKET:}"
_DELETE_CLUSTER="${DELETE_CLUSTER:-true}"
_ASM_MANAGED="${ASM_MANAGED:-false}"
_RELEASE_CHANNEL="${RELEASE_CHANNEL:-REGULAR}"
_SKIP_UNIT_TESTS="${SKIP_UNIT_TESTS:-true}"
_EXTRA_CERTS_URL="${_EXTRA_CERTS_URL:-}"

# Get Change-Id from commit. When in prow there is a merge commit after the CL
# commit so we list Change-Ids found in the previous 2 commits to find the CL
# commit.
change_id=$(git log -n2 | egrep '^\s*Change-Id:\s*\w+\s*$' | head -n1 | awk '{print tolower($2)}')
if [ -z "${change_id}" ]; then
    echo "Unable to find a Change-Id. Are you using Gerrit?"
    exit 1
fi
change_id_trunc=$(echo ${change_id} | awk '{print substr($1,0,37)}')

# Create Kf release.
git_sha=$(git rev-parse HEAD)
release_id="id-${change_id_trunc}-$(date +%s)"

gcloud builds submit . \
    --project ${_GCP_PROJECT_ID} \
    --config=cmd/generate-release/cloudbuild.yaml \
    --substitutions=_RELEASE_BUCKET=${_RELEASE_BUCKET},_GIT_SHA=${git_sha},_VERSION=${release_id}

# Run integration tests.
cluster_name=random
full_release_bucket=gs://${_RELEASE_BUCKET}/${release_id}

gcloud builds submit . \
    --project ${_GCP_PROJECT_ID} \
    --config=ci/cloudbuild/test.yaml \
    --substitutions="_CLOUDSDK_COMPUTE_ZONE=random,_CLOUDSDK_CONTAINER_CLUSTER=${cluster_name},_NODE_COUNT=6,_FULL_RELEASE_BUCKET=${full_release_bucket},_DELETE_CLUSTER=${_DELETE_CLUSTER},_MACHINE_TYPE=n1-highmem-4,_RELEASE_CHANNEL=${_RELEASE_CHANNEL},_SKIP_UNIT_TESTS=${_SKIP_UNIT_TESTS},_ASM_MANAGED=${_ASM_MANAGED}",_EXTRA_CERTS_URL=${_EXTRA_CERTS_URL}
