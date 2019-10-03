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

[[ -n "${KO_DOCKER_REPO-}" ]] || ( echo KO_DOCKER_REPO must be set; exit 1)
[[ -n "${SERVICE_ACCOUNT_JSON-}" ]] || ( echo SERVICE_ACCOUNT_JSON must be set; exit 1)
[[ -n "${GCP_PROJECT_ID-}" ]] || ( echo GCP_PROJECT_ID must be set; exit 1)

if [ "$#" -ne 1 ]; then
        echo "usage: $0 [OUTPUT PATH]"
        exit 1
fi

# Change to the project root directory
cd "${0%/*}"/..

# Directory where output artifacts will go
output=$1

# Login to gcloud
/bin/echo Authenticating to kubernetes...
sakey=`mktemp -t gcloud-key-XXXXXX`
set +x
/bin/echo "$SERVICE_ACCOUNT_JSON" > $sakey
set -x
gcloud auth activate-service-account --key-file $sakey
gcloud config set project "$GCP_PROJECT_ID"
gcloud -q auth configure-docker

# Process version
version=`cat version`
# Modify version number if this is a nightly build.
# Set NIGHTLY to any value to enable a nightly build.
[[ -n "${NIGHTLY-}" ]] && version="$version-nightly-`date +%F`"
echo $version > $output/version

# Modify version number to prepend git hash
hash=$(git rev-parse --short HEAD)
[ -z "$hash" ] && echo "failed to read hash" && exit 1
version="$version-$hash"

# Temp dir for temp GOPATH
tmpgo=`mktemp -d -t go-XXXXXXXXXX`
# Environment Variables for go build
export GOPATH=$tmpgo
export GOPROXY=https://proxy.golang.org
export GOSUMDB=sum.golang.org
export GO111MODULE=on
export CGO_ENABLED=0

#########################
# Generate Release YAML #
#########################

# ko requires a proper go path and deps to be vendored
# TODO remove this once https://github.com/google/ko/issues/7 is
# resolved.
go mod vendor
mkdir -p $GOPATH/src/github.com/google/
ln -s $PWD $GOPATH/src/github.com/google/kf
cd $GOPATH/src/github.com/google/kf

# ko resolve
# This publishes the images to KO_DOCKER_REPO and writes the yaml to
# stdout. $version is substited in the yaml and the result is written
# to a file.
ko resolve --filename config | sed "s/VERSION_PLACEHOLDER/$version/" > ${output}/release.yaml

###################
# Generate kf CLI #
###################
mkdir ${output}/bin

# Build the binaries
for os in $(echo linux darwin windows); do
  destination=${output}/bin/kf-${os}
  if [ ${os} = "windows" ]; then
    # Windows only executes things with the .exe extension
    destination=${destination}.exe
  fi

  # Build
  GOOS=${os} go build -o ${destination} --ldflags "-X 'github.com/google/kf/pkg/kf/commands.Version=${version}'" ./cmd/kf
done
