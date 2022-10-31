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


set -eux

cd "${0%/*}"/..

if [ "${VERSION}" = "" ]; then
  echo "VERSION is required (e.g., v1.0.3)"
  exit 1
fi

# Get doc version, doc version is the numeric value of the version,
# e.g. VERSION=v2.3.0, DOC_VERSION=2.3, DOC_VERSION is used at doc URL path.
IFS='.'
read -ra VERSION_NUMBERS <<< "${VERSION}"
MAJOR="${VERSION_NUMBERS[0]}"
MINOR="${VERSION_NUMBERS[1]}"
DOC_VERSION="${MAJOR:1}.${MINOR}"
echo "DOC_VERSION:${DOC_VERSION}"

cli_output_dir="${PWD}/docs/content/en/docs/v${DOC_VERSION}/cli/generated"
troubleshoot_output_dir="${PWD}/docs/content/en/docs/v${DOC_VERSION}/troubleshooting/generated"

temp_dir=$(mktemp -d)
finish() {
  rm -rf "${temp_dir}"
}
trap finish EXIT

# Generate docs
go run ./pkg/kf/internal/tools/command-doc-generator/command-doc-generator.go "${temp_dir}" "${DOC_VERSION}"

# This can get a bit noisy, turn off verbose mode.
set +x
pushd "${temp_dir}"
for file in *.md ; do
  # Replace '_' with '-' in names
  new_name="${file//_/-}"
  if [ "${new_name}" != "${file}" ] && [ "${file}" != "_index.md" ]; then
    mv "${file}" "${new_name}"
  elif [ "${file}" == "kf.md" ]; then
    # We don't need this file.
    rm "${file}"
  fi
done
popd
set -x

# Remove old generated files then add the new ones.
pushd "${cli_output_dir}"
rm -f ./*.md
cp -r "${temp_dir}/"* .
popd

# Set up runbooks
rm -r "${temp_dir}/"

go run ./pkg/kf/internal/tools/runbook-generator/main.go "${temp_dir}" "${DOC_VERSION}"

# Remove old generated files then add the new ones.
pushd "${troubleshoot_output_dir}"
rm -f ./*.md
cp -r "${temp_dir}/"* .
popd
