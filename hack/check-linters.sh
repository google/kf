#!/usr/bin/env sh

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

set -eu

# gofmt -s -d
GOFMT_DIFF=$(IFS=$'\n' gofmt -s -d $( find . -type f -name '*.go' ) )
if [ -n "${GOFMT_DIFF}" ]; then
    echo "${GOFMT_DIFF}"
    echo
    echo "The go source files aren't gofmt formatted."
    exit 1
fi

go vet ./...

# Checking for misspelled words
GO111MODULE=off go get -u github.com/client9/misspell/cmd/misspell
misspell -error .
