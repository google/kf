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

set -eu

version=${VERSION:-latest}
publish=${PUBLISH:-false}

rootpath=$(cd $(dirname $0)/.. && pwd)
dir=${rootpath}/samples/buildpacks/stacks/bionic
KF_REGISTRY=$1

base_image=${KF_REGISTRY}/base:${version}
run_image=${KF_REGISTRY}/run:${version}
build_image=${KF_REGISTRY}/build:${version}
docker build -t "${base_image}" "$dir/base"
docker build --build-arg "base_image=${base_image}" -t "${build_image}"  "$dir/build"
docker build --build-arg "base_image=${base_image}" -t "${run_image}" "$dir/run"

if [ $publish == "true" ]; then
    echo "Publishing these images..."
    for image in "${base_image}" "${run_image}" "${build_image}"; do
      docker push ${image}
    done
else
    echo "To publish these images:"
    for image in "${base_image}" "${run_image}" "${build_image}"; do
      echo "  docker push ${image}"
    done
fi
