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

# Login to gcloud
set +x
/bin/echo "$SERVICE_ACCOUNT_JSON" > key.json
set -x
/bin/echo Authenticating to kubernetes...
gcloud auth activate-service-account --key-file key.json
gcloud config set project "$GCP_PROJECT_ID"
gcloud -q auth configure-docker

# Used to suffix the artifacts
current_time=$(date +%s)

#########################
# Generate Release YAML #
#########################

# Environment Variables for go build
export GOPATH=/go
export GOPROXY=https://proxy.golang.org
export GOSUMDB=sum.golang.org
export GO111MODULE=on
export CGO_ENABLED=0

# ko requires a proper go path and deps to be vendored
# TODO remove this once https://github.com/google/ko/issues/7 is
# resolved.
go mod vendor
mkdir -p $GOPATH/src/github.com/google/
ln -s $PWD $GOPATH/src/github.com/google/kf
cd $GOPATH/src/github.com/google/kf

release_name=release-${current_time}.yaml
# ko resolve
# This publishes the images to KO_DOCKER_REPO and writes the yaml to
# stdout.
ko resolve --filename config > /tmp/${release_name}

# save release to a bucket
gsutil cp /tmp/${release_name} ${RELEASE_BUCKET}/${release_name}

# Make this the latest
gsutil cp ${RELEASE_BUCKET}/${release_name} ${RELEASE_BUCKET}/release-latest.yaml

###################
# Generate kf CLI #
###################

hash=$(git rev-parse HEAD)
[ -z "$hash" ] && echo "failed to read hash" && exit 1

# Build and upload the binaries
for os in $(echo linux darwin windows); do
  destination=kf-${os}-${current_time}
  latest_destination=kf-${os}-latest
  if [ ${os} = "windows" ]; then
    # Windows only executes things with the .exe extension
    destination=${destination}.exe
    latest_destination=${latest_destination}.exe
  fi

  # Build
  GOOS=${os} go build -o /tmp/${destination} "-X github.com/google/kf/pkg/kf/commands.Version=${hash}" ./cmd/kf

  # Upload
  gsutil cp /tmp/${destination} ${CLI_RELEASE_BUCKET}/${destination}

  # Make this the latest
  gsutil cp ${CLI_RELEASE_BUCKET}/${destination} ${CLI_RELEASE_BUCKET}/${latest_destination}
done
