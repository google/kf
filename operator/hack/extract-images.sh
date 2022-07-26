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


source gbash.sh || exit

version=$(git rev-parse --abbrev-ref HEAD | cut -d- -f2)
DEFINE_array operand_name --type=enum --enum="eventing,knative-serving,metrics-collector,eventing-post-install" "eventing,knative-serving/$version,metrics-collector,eventing-post-install" "The operand name."

main() {
    for operand_name in "${FLAGS_operand_name[@]}"; do
      grep -io "[^\b]*.gcr.*$" "cmd/manager/kodata/${operand_name}/"* \
        | sed -e 's/[^\b]*.gcr.io/{{.AddonsImageRegistry}}/g' \
        | awk -F'[/@]' '{print $(NF - 2)"."$(NF - 1)": '\''"$0"'\''"}' \
        | sort -u
    done
}

gbash::main "$@"
