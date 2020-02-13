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

set -eu

cd "${0%/*}"/..

# ko requires a proper go path and deps to be vendored
# TODO remove this once https://github.com/google/ko/issues/7 is
# resolved.
echo "Vendoring from: `pwd`"
go mod vendor

readonly kfpath="$GOPATH/src/github.com/google/kf"

echo "Checking that $kfpath exists"
if [[ ! -d "$kfpath" ]]; then
  echo "Linking $kfpath"
  mkdir -p $GOPATH/src/github.com/google/
  ln -s $PWD $kfpath
fi

pushd $kfpath
ko apply -f config
popd
