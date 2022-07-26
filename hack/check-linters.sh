#!/usr/bin/env bash

# Copyright 2019 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -eux

cd "${0%/*}"/..

# Use mapfile to store an array of the go packages.
mapfile -t packages < <(find . -type f -name '*.go' | grep -v \./vendor/ | grep -v \./third_party | grep -v \./first_party)

# gofmt -s -d
GOFMT_DIFF=$(gofmt -s -d "${packages[@]}")
if [ -n "${GOFMT_DIFF}" ]; then
    echo "${GOFMT_DIFF}"
    echo
    echo "The go source files aren't gofmt formatted."
    exit 1
fi

go list ./... | grep -v ^github.com/google/kf/v2/third_party | grep -v ^github.com/google/kf/v2/vendor | grep -v ^github.com/google/kf/v2/first_party | xargs go vet

# Checking for misspelled words
GO111MODULE=off go get -u github.com/client9/misspell/cmd/misspell

# ignore vendor directory
find . -type f -name '*.go' | grep -v ./vendor/ | grep -v ./third_party/ | grep -v ./first_party/ | xargs -n1 -P 20 "$(go env GOPATH)/bin/misspell" -error
