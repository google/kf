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

export GO111MODULE=on
export GOPROXY=https://proxy.golang.org
export GOSUMDB=sum.golang.org

# Change to the project root directory
cd "${0%/*}"/..

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
