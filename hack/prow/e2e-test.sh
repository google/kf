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

_GCP_PROJECT_ID="${GCP_PROJECT_ID:-}"
_RELEASE_BUCKET="${RELEASE_BUCKET:-}"
_DELETE_CLUSTER="${DELETE_CLUSTER:-true}"
_ASM_MANAGED="${ASM_MANAGED:-false}"
_RELEASE_CHANNEL="${RELEASE_CHANNEL:-REGULAR}"
_SKIP_UNIT_TESTS="${SKIP_UNIT_TESTS:-true}"
_EXTRA_CERTS_URL="${_EXTRA_CERTS_URL:-}"

if ! [ -x "$(command -v jq)" ]; then
  apk add --update --no-cache jq
fi

# Configur result exporting
_EXPORT_BUCKET="${EXPORT_BUCKET:-}"
_EXPORT_JOB_NAME="${EXPORT_JOB_NAME:-}"
_EXPORT_REPO="${EXPORT_REPO:-google/kf}"

# Create Kf release
git_sha=${COMMIT_SHA:-$(git rev-parse HEAD)}
release_id="$(date +%s)-$(git rev-parse --short $git_sha)"

# Determine the path to export results to and store the latest release ID
export_path=gs://${_EXPORT_BUCKET}/logs/${_EXPORT_JOB_NAME}/$release_id
[ ! -z "${_EXPORT_BUCKET}" ] && echo $release_id | gsutil cp - gs://${_EXPORT_BUCKET}/logs/${_EXPORT_JOB_NAME}/latest-build.txt

# Write 'started.json' to report start of job to testgrid
[ ! -z "${_EXPORT_BUCKET}" ] && jq -n \
    --argjson timestamp "$(date +%s)" \
    --argjson repo "$( \
        jq -n \
            --arg "${_EXPORT_REPO}" $git_sha \
            '$ARGS.named' \
    )" \
    '$ARGS.named' | gsutil cp - ${export_path}/started.json

# Install exit hook to write 'finished.json' to report results to testgrid
RESULT="FAILURE"
function finish {
    echo "Integration test status: $RESULT"
    if [ "$RESULT" = "SUCCESS" ]; then RESULT_PASSED="true"; else RESULT_PASSED="false"; fi
    if [ ! -z "${_EXPORT_BUCKET}" ]; then
        jq -n \
              --argjson timestamp "$(date +%s)" \
              --arg result "${RESULT}" \
              --argjson passed "${RESULT_PASSED}" \
              --argjson metadata "$( \
                  jq -n \
                      --arg "Commit" $git_sha \
                      --arg "Build" "$BUILD_ID" \
                      '$ARGS.named' \
              )" \
              '$ARGS.named' | gsutil cp - ${export_path}/finished.json
        gsutil cp ./build-log.txt ${export_path}/build-log.txt
    fi
}
trap finish EXIT

# Run the build
gcloud builds submit . \
    --project ${_GCP_PROJECT_ID} \
    --config=cmd/generate-release/cloudbuild.yaml \
    --substitutions=_RELEASE_BUCKET=${_RELEASE_BUCKET},_GIT_SHA=${git_sha},_VERSION=${release_id} \
    | tee -a ./build-log.txt

# Run integration tests.
cluster_name=random
full_release_bucket=gs://${_RELEASE_BUCKET}/${release_id}

gcloud builds submit . \
    --project ${_GCP_PROJECT_ID} \
    --config=ci/cloudbuild/test.yaml \
    --substitutions="_CLOUDSDK_COMPUTE_ZONE=random,_CLOUDSDK_CONTAINER_CLUSTER=${cluster_name},_NODE_COUNT=6,_FULL_RELEASE_BUCKET=${full_release_bucket},_DELETE_CLUSTER=${_DELETE_CLUSTER},_MACHINE_TYPE=n1-highmem-4,_RELEASE_CHANNEL=${_RELEASE_CHANNEL},_SKIP_UNIT_TESTS=${_SKIP_UNIT_TESTS},_ASM_MANAGED=${_ASM_MANAGED}",_EXTRA_CERTS_URL=${_EXTRA_CERTS_URL} \
    | tee -a ./build-log.txt

# Update result to success if we reach the end, elsewise we'll report a failure
RESULT="SUCCESS"