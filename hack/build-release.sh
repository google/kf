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

if [ "$#" -ne 1 ]; then
        echo "usage: $0 [OUTPUT PATH]"
        exit 1
fi

output=$1

# Change to the project root directory
cd "${0%/*}"/..

# Login to gcloud
set +x
/bin/echo "$SERVICE_ACCOUNT_JSON" > key.json
set -x
/bin/echo Authenticating to kubernetes...
gcloud auth activate-service-account --key-file key.json
gcloud config set project "$GCP_PROJECT_ID"
gcloud -q auth configure-docker

# Environment Variables for go build
export GOPATH=/go
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
# stdout.
ko resolve --filename config > ${output}/release.yaml

###################
# Generate kf CLI #
###################

mkdir ${output}/bin

hash=$(git rev-parse HEAD)
[ -z "$hash" ] && echo "failed to read hash" && exit 1

# Build the binaries
for os in $(echo linux darwin windows); do
  destination=${output}/bin/kf-${os}
  if [ ${os} = "windows" ]; then
    # Windows only executes things with the .exe extension
    destination=${destination}.exe
  fi

  # Build
  GOOS=${os} go build -o ${destination} --ldflags "-X github.com/google/kf/pkg/kf/commands.Version=${hash}" ./cmd/kf
done
