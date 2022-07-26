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

temp_dir=$(mktemp -d)
finish() {
  rm -rf "${temp_dir}"
}
trap finish EXIT

# Generate docs
go run ./pkg/kf/internal/tools/command-doc-generator/command-doc-generator.go "${temp_dir}" "${DOC_VERSION}"

# Build a new _index.md file
# We want to use the kf.md file with a different header.
go run ./pkg/kf/internal/tools/command-doc-generator/command-doc-generator.go --kf-only "${temp_dir}/index.md" "${DOC_VERSION}"

# Build a new _book.yaml file
go run ./pkg/kf/internal/tools/command-doc-generator/command-doc-generator.go --book-only "${temp_dir}/_toc.yaml" "${DOC_VERSION}"

# Build a new dependency_matrix.html file.
cat << EOF > "${temp_dir}/dependency-matrix-${VERSION}.md"
{% include "migrate/kf/docs/${DOC_VERSION}/_local_variables.html" %}
{% include "_shared/apis/_clientlib_variables.html" %}
{% include "_shared/apis/console/_local_variables.html" %}
{% include "cloud/_shared/_cloud_shared_files.html" %}

EOF
go run ./cmd/kf dependencies matrix >> "${temp_dir}/dependency-matrix-${VERSION}.md"

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

# Setup a new CitC
clientdir="$(p4 g4d -f -- "${CITC:=kf-cli-docs-$(date +%s)}")"
pushd "${clientdir}"
output_dir=${clientdir}/googledata/devsite/site-cloud/en/migrate/kf/docs/${DOC_VERSION}/cli/
mkdir -p "$output_dir"
popd

# CitC isn't a normal filedrive, so visiting it instead of calling it from afar
# helps it work more reliably.
pushd "${clientdir}/googledata/devsite/site-cloud/en/migrate/kf/docs/${DOC_VERSION}/"

mv "${temp_dir}/dependency-matrix-${VERSION}.md" "./_dependency-matrix-${VERSION}.md"

popd

# CitC isn't a normal filedrive, so visiting it instead of calling it from afar
# helps it work more reliably.
pushd "${output_dir}"

# Remove everything from the commands directory.
# We do this so if we ever delete a command, the doc won't be stranded.
rm -f ./*.md

cp -r "${temp_dir}/"* .

popd

# Set up runbooks
rm -r "${temp_dir}/"

go run ./pkg/kf/internal/tools/runbook-generator/main.go "${temp_dir}" "${DOC_VERSION}"

pushd "${clientdir}/googledata/devsite/site-cloud/en/migrate/kf/docs/${DOC_VERSION}/reference"
cp -r "${temp_dir}/"* .
popd

# Check to see if a CL has already been created for this CitC
pushd "${clientdir}"
cl=$(p4 -F %change0% p)
if [[ "${cl}" == "" ]]; then
  # Create a CL
  p4 change --desc "Regenerate Kf CLI command docs and dependency matrix"
  cl=$(p4 -F %change0% p)
  echo "Next steps:"
  echo "cl/${cl}"
  echo "To stage to devsite:"
  echo "/google/data/ro/projects/devsite/devsite2 stage --cl=${cl}"
fi
popd
