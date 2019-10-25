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

cd $(dirname $(go env GOMOD))

echo "Vetting"
go vet ./...

echo "Formatting"
gofmt -s -w $( find . -type f -name '*.go' | grep -v \./vendor/)

echo "Checking spelling"
GO111MODULE=off go get -u github.com/client9/misspell/cmd/misspell
find . -type f -name '*.go' | grep -v ./vendor/ | xargs -n1 -P 20 $(go env GOPATH)/bin/misspell -error

echo "Generating updates"
go generate ./...
go mod tidy

echo "Generating code-generator packages"
./hack/update-codegen.sh

echo "Updating license"
./hack/check-vendor-license.sh

echo "Updating go.sum"
go mod tidy
