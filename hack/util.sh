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


require_env() {
  local name=$1
  local value
  value="$(eval echo '$'"$name")"

  if [ "$value" == '' ]; then
    echo "environment variable $name must be set"
    exit 1
  fi
}

# retry will run the command, if it fails, it will sleep and then try again:
# args:
# 1: command: The command to run. (required)
# 2: retry_count: The number of times to attempt the command. (defaults to 3)
# 3: sleep_seconds: The number of seconds to sleep after a failure. (defaults to 1)
function retry() {
    command=${1:-x}
    retry_count=${2:-3}
    sleep_seconds=${3:-1}

    if [ "${command}" = "x" ]; then
        echo "command missing"
        exit 1
    fi

    n=0
    until [ "$n" -ge ${retry_count} ]
    do
       ${command} && break
       echo "failed..."
       n=$((n+1))
       sleep ${sleep_seconds}
    done

    if [ ${n} = ${retry_count} ];then
        echo "\"${command}\" failed too many times (${retry_count})"
        exit 1
    fi
}

# Helper function to gracefully fail a script by printing
# the error message and returning with exit code.
function fail_script() {
  [[ -n "$1" ]] && echo "ERROR: $1. Failing script."
  exit 1
}

# Helper method to apply operator yaml
# This won't be needed after b/181898563 got resolved
function apply_yaml() {
  local file=$1
  local retry
  local max_retries=3
  for (( retry=0; retry<max_retries; retry++)) do
    if kubectl apply -f "${file}"; then
      echo "Successfully applied ${file}"
      return
    else
      sleep 5s
    fi
  done

  fail_script "Failed applying ${file}"
}

function check_env() {
  local env=$1
  if [[ ! -v "$env" ]]; then
    echo "$env not set"
    exit 1
  fi
}

# version gets a version hash from the current git commit. If the directory is
# dirty it appends a hash of the git porcelain.
version() {
    local version
    version="$(git rev-parse --short HEAD)"
    if [ -n "$(git status --porcelain)" ]; then
        version="${version}-$(git status --porcelain \
            | md5sum \
            | awk '{ print substr($1, 1, 8)  }')"
    fi
    echo "$version"
}
