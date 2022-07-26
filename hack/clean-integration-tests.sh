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

# Use mapfile to store an array of integration namespaces.
mapfile -t spaces < <(kubectl get spaces | grep -e 'integration-test-[0-9]*' | cut -d' ' -f1)
if [[ ${#spaces[@]} -ne 0 ]]; then
  echo "The following spaces will be deleted:"
  echo "${spaces[*]}"
  echo
  echo "Press anything to continue"
  read -r
  echo "${spaces[@]}" | xargs -n1 kubectl delete space
fi

# Use mapfile to store an array of leftover yml files.
mapfile -t files1 < <(ls samples/apps/spring-music/manifest-integration-*)
mapfile -t files2 < <(git status | grep -e 'manifest-integration-[0-9]*.yml' | sed 's/\s//g')
files=("${files1[@]}" "${files2[@]}")
if [[ ${#files[@]} -ne 0 ]]; then
  echo "The following files will be deleted:"
  echo "${files[*]}"
  echo
  echo "Press anything to continue"
  read -r
  echo "${files[@]}" | xargs -n1 rm -f
fi
