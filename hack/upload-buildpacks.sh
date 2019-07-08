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

scriptpath=$(cd $(dirname $0)/.. && pwd)

# Set the container registry
# If KF_REGISTRY is populated then use that otherwise try setting it to gcr
function setregistry {
 if [ -z ${KF_REGISTRY:-} ]; then
    proj_id=$(gcloud config get-value project)
    if [ -z $proj_id ]; then
       echo "could not set container registry..."
       exit 1
    else
       KF_REGISTRY="gcr.io/$proj_id"
    fi
 fi
}
setregistry

# Upload stacks
PUBLISH=true $scriptpath/hack/upload-buildpacks-stack.sh $KF_REGISTRY
samples=$(realpath $scriptpath/samples)
builder_config=$samples/buildpacks/builder/builder.toml
builder_image="$KF_REGISTRY/buildpack-builder-$RANDOM"

# Create a temp directory so we can manipulate the builder.toml. We can't
# simply use a temp file "<(...)" because the builder.toml has relative paths
# to the buildpacks.
temp_dir=$(mktemp -d)
function finish {
  rm -rf "$temp_dir"
}
trap finish EXIT
cp -r $samples/buildpacks/builder/* $temp_dir

# Fill in our container registry for the builder
builder_toml=$temp_dir/builder.toml
cat $builder_toml | sed "s|REPLACE_WITH_REGISTRY|$KF_REGISTRY|g" > $builder_toml.new

echo "building builder from $builder_config..."
pack create-builder \
    $builder_image \
    --publish \
    --builder-config \
    $builder_toml.new

echo "updating buildpack ClusterBuildTemplate..."
cluster_build_template='
{
    "apiVersion": "build.knative.dev/v1alpha1",
    "kind": "ClusterBuildTemplate",
    "metadata": {
        "name": "buildpack"
    },
    "spec": {
        "parameters": [
            {
                "description": "The image you wish to create. For example, \"repo/example\", or \"example.com/repo/image\"",
                "name": "IMAGE"
            },
            {
                "default": "packs/run:v3alpha2",
                "description": "The run image buildpacks will use as the base for IMAGE.",
                "name": "RUN_IMAGE"
            },
            {
                "default": "XXX_REPLACE_XXX",
                "description": "The builder image (must include v3 lifecycle and compatible buildpacks).",
                "name": "BUILDER_IMAGE"
            },
            {
                "default": "true",
                "description": "Use Docker credential helpers for Googles GCR, Amazons ECR, or Microsofts ACR.",
                "name": "USE_CRED_HELPERS"
            },
            {
                "default": "empty-dir",
                "description": "The name of the persistent app cache volume",
                "name": "CACHE"
            },
            {
                "default": "1000",
                "description": "The user ID of the builder image user",
                "name": "USER_ID"
            },
            {
                "default": "1000",
                "description": "The group ID of the builder image user",
                "name": "GROUP_ID"
            },
            {
                "default": "",
                "description": "When set, skip the detect step and use the given buildpack.",
                "name": "BUILDPACK"
            }
        ],
        "steps": [
            {
                "args": [
                    "-c",
                    "chown -R \"${USER_ID}:${GROUP_ID}\" \"/builder/home\" \u0026\u0026\n\t\t\t\t\t\t chown -R \"${USER_ID}:${GROUP_ID}\" /layers \u0026\u0026\n\t\t\t\t\t\t chown -R \"${USER_ID}:${GROUP_ID}\" /workspace"
                ],
                "command": [
                    "/bin/sh"
                ],
                "image": "alpine",
                "imagePullPolicy": "Always",
                "name": "prepare",
                "resources": {},
                "volumeMounts": [
                    {
                        "mountPath": "/layers",
                        "name": "${CACHE}"
                    }
                ]
            },
            {
                "args": [
                    "-c",
                    "if [[ -z \"${BUILDPACK}\" ]]; then\n  /lifecycle/detector \\\n  -app=/workspace \\\n  -group=/layers/group.toml \\\n  -plan=/layers/plan.toml\nelse\ntouch /layers/plan.toml\ncat \u003c\u003cEOF \u003e /layers/group.toml\n[[buildpacks]]\n  id = \"${BUILDPACK}\"\n  version = \"latest\"\nEOF\nfi"
                ],
                "command": [
                    "/bin/bash"
                ],
                "image": "${BUILDER_IMAGE}",
                "imagePullPolicy": "Always",
                "name": "detect",
                "resources": {},
                "volumeMounts": [
                    {
                        "mountPath": "/layers",
                        "name": "${CACHE}"
                    }
                ]
            },
            {
                "args": [
                    "-layers=/layers",
                    "-helpers=${USE_CRED_HELPERS}",
                    "-group=/layers/group.toml",
                    "${IMAGE}"
                ],
                "command": [
                    "/lifecycle/analyzer"
                ],
                "image": "${BUILDER_IMAGE}",
                "imagePullPolicy": "Always",
                "name": "analyze",
                "resources": {},
                "volumeMounts": [
                    {
                        "mountPath": "/layers",
                        "name": "${CACHE}"
                    }
                ]
            },
            {
                "args": [
                    "-layers=/layers",
                    "-app=/workspace",
                    "-group=/layers/group.toml",
                    "-plan=/layers/plan.toml"
                ],
                "command": [
                    "/lifecycle/builder"
                ],
                "image": "${BUILDER_IMAGE}",
                "imagePullPolicy": "Always",
                "name": "build",
                "resources": {},
                "volumeMounts": [
                    {
                        "mountPath": "/layers",
                        "name": "${CACHE}"
                    }
                ]
            },
            {
                "args": [
                    "-layers=/layers",
                    "-helpers=${USE_CRED_HELPERS}",
                    "-app=/workspace",
                    "-image=${RUN_IMAGE}",
                    "-group=/layers/group.toml",
                    "${IMAGE}"
                ],
                "command": [
                    "/lifecycle/exporter"
                ],
                "image": "${BUILDER_IMAGE}",
                "imagePullPolicy": "Always",
                "name": "export",
                "resources": {},
                "volumeMounts": [
                    {
                        "mountPath": "/layers",
                        "name": "${CACHE}"
                    }
                ]
            }
        ],
        "volumes": [
            {
                "name": "empty-dir"
            }
        ]
    }
}
'

echo $cluster_build_template | sed "s|XXX_REPLACE_XXX|$builder_image|" | kubectl apply -f -
