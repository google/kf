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

# Change to the project root directory
cd "${0%/*}"/..

# Used to suffix the artifacts
current_time=$(date +%s)

# Build artifacts
# NOTE: This will login to gcloud for us.
tmp_dir=$(mktemp -d)
./hack/build-release.sh ${tmp_dir}

# save release to a bucket
release_name=release-${current_time}.yaml
gsutil cp ${tmp_dir}/release.yaml ${RELEASE_BUCKET}/${release_name}

# Make this the latest
gsutil cp ${RELEASE_BUCKET}/${release_name} ${RELEASE_BUCKET}/release-latest.yaml

# Upload the binaries
for cli in $(ls ${tmp_dir}/bin); do
  # Extract file extensions (e.g., .exe)
  filename=$(basename -- "${cli}")
  extension="${filename##*.}"
  filename="${filename%.*}"

  destination=${filename}-${current_time}
  latest_destination=${filename}-latest

  # Check to see if it has a file extension
  if [ "${extension}" != "${filename}" ]; then
      destination=${destination}.${extension}
      latest_destination=${latest_destination}.${extension}
  fi

  # Upload
  gsutil cp ${tmp_dir}/bin/${cli} ${CLI_RELEASE_BUCKET}/${destination}

  # Make this the latest
  gsutil cp ${CLI_RELEASE_BUCKET}/${destination} ${CLI_RELEASE_BUCKET}/${latest_destination}
done
