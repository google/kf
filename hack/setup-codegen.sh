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
# retrieve the code-generator scripts and bins

set -xeu

cd $(go env GOPATH)

packages="
github.com/golang/mock/mockgen/model
k8s.io/kubernetes/pkg/apis/core/...
k8s.io/code-generator/...
k8s.io/apimachinery/...
k8s.io/api/core/v1/...
k8s.io/client-go/...
github.com/knative/serving/pkg/apis/...
github.com/knative/build/pkg/apis/...
knative.dev/pkg/...
github.com/knative/serving/apis/...
github.com/poy/service-catalog/pkg/apis/...
"

for p in $packages; do
  go get -u $p
done
