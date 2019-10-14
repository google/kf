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

# This script is used by the CI to check if the code is gofmt formatted.
# This script is used by the CI to check if 'go generate ./...' is up to date.

set -eu

go install github.com/google/wire/cmd/wire
go install github.com/golang/mock/mockgen
export PATH="$PATH:$(go env GOPATH)/bin"
go generate ./...
go mod tidy

if [ ! -z "$(git status --porcelain)" ]; then
    git status
    echo
    echo "The generated files aren't up to date."
    echo "Update them with the 'go generate ./...' command."
    git --no-pager diff
    exit 1
fi
