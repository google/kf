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

export GO111MODULE=on

cd $(dirname "$0")

project_id=${PROJECT_ID:""}
if [ -z "$project_id" ]; then
    project_id="$(gcloud config get-value project)"
fi

write_dockerfile () {
    local name=$1
    local from_alpine=$2
    if [ "$from_alpine" = true ]; then
        run_from=alpine
    else
        run_from=gcr.io/$project_id/kf-prefix-output
    fi

	cat << EOF >> Dockerfile
# build stage
FROM golang:1.12 AS builder
ADD . /src
ENV CGO_ENABLED=0
ENV GO111MODULE=on
RUN cd /src && go build -o $name --mod=vendor

# final stage
FROM $run_from
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
WORKDIR /app
COPY --from=builder /src/$name /app/
EOF

    if [ ! "$from_alpine" = true ]; then
	cat << EOF >> Dockerfile
ENTRYPOINT ["/app/prefix-output", "/app/$name"]
CMD []
EOF
    fi
}

add_vendor () {
    # Fetch path to kf dir
    local dir=$1
    local tmp_mod=$(mktemp)

    # Remove replace kf line
    cat go.mod | egrep -v 'replace\s+github.com/GoogleCloudPlatform/kf\s+' > "$tmp_mod"
    mv $tmp_mod go.mod

    # Write new replace line that is relative to tmp path
    echo "replace github.com/GoogleCloudPlatform/kf => $(realpath --relative-to="$PWD" "$dir")" >> go.mod

    # Add vendor directory
    go mod vendor
}

build_and_push () {
    local dir=$(pwd)/$1
    local use_run_alpine=$2
    local name=$(basename $dir)
    local root_dir=$(git rev-parse --show-toplevel)

    # Lets do everything in a temp directory so we don't blow anything away
    # while adding vendoring and a Dockerfile.
    local tmp_dir=$(mktemp -d)
    trap "rm -rf $tmp_dir" EXIT
    cp -r $dir/* $tmp_dir

    pushd $tmp_dir &> /dev/null
    add_vendor $root_dir
    write_dockerfile $name $use_run_alpine
    docker build -t gcr.io/${project_id}/kf-${name} .
    docker push gcr.io/${project_id}/kf-${name}
    popd &> /dev/null
}

build_and_push ./cmd/prefix-output true

for dir in $(ls -d ./cmd/*/ | grep -v generator | grep -v prefix-output); do
    ( build_and_push $dir false ) &
done

wait

echo "apply CRD and ClusterBindingRoles"
go run ./cmd/generator --container-registry="gcr.io/${project_id}" | kubectl apply -f -
