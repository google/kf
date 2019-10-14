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

set -eux

source "`dirname $0`/util.sh"

require_env SERVICE_ACCOUNT_JSON
require_env GCP_PROJECT_ID
require_env RELEASE_BUCKET

if [ "$#" -ne 1 ]; then
        echo "usage: $0 [RELEASE ARTIFACTS PATH]"
        exit 1
fi

# Directory where output artifacts will go
release_dir=$1

# Login to gcloud
echo Authenticating to kubernetes...
sakey=`mktemp -t gcloud-key-XXXXXX`
set +x
echo "$SERVICE_ACCOUNT_JSON" > $sakey
set -x
gcloud auth activate-service-account --key-file $sakey
gcloud config set project "$GCP_PROJECT_ID"

# Parse the release version from a file
release_version=`cat $release_dir/version`
prefix="$release_version/"
prefix_latest="latest/"

# Modify the prefix if this is a nightly
if [[ $release_version == *"nightly"* ]]; then
  prefix="nightly/$prefix"
  prefix_latest="nightly/$prefix_latest"
fi

# Upload release
gsutil -m cp -a public-read -r $release_dir/* $RELEASE_BUCKET/$prefix

# Upload latest
gsutil -m cp -a public-read -r $release_dir/* $RELEASE_BUCKET/$prefix_latest
