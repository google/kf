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

set -eu

readonly target=${1:?Error: Please supply a target}
shift
readonly pipeline=${1:?Error: Please supply a pr number}
shift
readonly branch=${1:?Error: Please supply a git branch}
shift

set -x

config=ci/concourse/pipelines/pr-pipeline.yml

fly -t $target set-pipeline -p $pipeline -c $config -v pr_number=$pipeline -v git_branch=$branch