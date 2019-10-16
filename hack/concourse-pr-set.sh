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

set -e

usage="Usage: $0 -t fly-target -p pull-request -u git-uri -b git-branch"

while getopts t:p:u:b: o
do case "$o" in
  t)  target="$OPTARG";;
  p)  pipeline="$OPTARG";;
  u)  uri="$OPTARG";;
  b)  branch="$OPTARG";;
  [?])  echo $usage >&2 && exit 1;;
  esac
done

[ "$target" != "" ] || (echo "Error: Please supply a target. $usage" >&2 && exit 1)
[ "$pipeline" != "" ] || (echo "Error: Please supply a pipeline. $usage" >&2 && exit 1)
[ "$uri" != "" ] || (echo "Error: Please supply a uri. $usage" >&2 && exit 1)
[ "$branch" != "" ] || (echo "Error: Please supply a branch. $usage" >&2 && exit 1)

set -x

config=ci/concourse/pipelines/pr-pipeline.yml

fly -t $target set-pipeline -p $pipeline -c $config -v pr_number=$pipeline -v git_uri=$uri -v git_branch=$branch
