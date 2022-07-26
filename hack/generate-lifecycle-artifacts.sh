#!/usr/bin/env bash
# Copyright 2022 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -euo pipefail

cd "${0%/*}"/..

# Generate v2 lifecycle image artifacts. Compiled the builder and launcher and
# places them in the kodata folder.
pushd ./third_party/forked/v2-buildpack-lifecycle
go build -o ./installer/kodata/builder ./builder
go build -o ./installer/kodata/launcher ./launcher
popd

# Place these in the vendor directory as well. ko can pull from either, so we'll
# copy to both.
cp -r \
    ./third_party/forked/v2-buildpack-lifecycle/installer/kodata \
    vendor/code.cloudfoundry.org/buildpackapplifecycle/installer/kodata
