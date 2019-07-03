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
github.com/knative/serving/pkg/apis/...
github.com/knative/build/pkg/apis/...
github.com/poy/service-catalog/pkg/apis/...
k8s.io/api/...
k8s.io/apimachinery/...
k8s.io/client-go/...
k8s.io/code-generator/...
knative.dev/pkg/...
"

for p in $packages; do
  go get -u $p
done

cd $(go env GOPATH)
mkdir -p src/github.com/knative/
cd src/github.com/knative
if [ ! -d "pkg" ]; then
  git clone https://github.com/knative/pkg
fi

cd pkg
if [ "$(git remote | grep poy)" = "" ]; then
  git remote add poy https://github.com/poy/knative-pkg
fi

git fetch poy
git checkout poy/poy-fix

echo ready to run codegen
