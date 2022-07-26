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

source ./hack/util.sh

check_env SPACE_CONTAINER_REGISTRY && check_env SERVICE_ACCOUNT && check_env PROJECT_ID && check_env DOMAIN
if [ $? -ne 0 ]; then
  exit 1
fi

# Get Kf yaml
VERSION="${VERSION:-$(version)}"

# This is necessary as of ko v0.6.0. They updated the default base image to
# use nonroot and therefore broke our Build pipeline which currently requires
# root. This may not be necessary after b/160025670 is fixed.
export KO_DEFAULTBASEIMAGE="gcr.io/distroless/static:latest"
export KO_DOCKER_REPO="${KO_DOCKER_REPO:-gcr.io/$(gcloud config get-value project)}"

temp_dir=$(mktemp -d)
trap 'rm -rf $temp_dir' EXIT
kf_release="${temp_dir}/kf.yaml"
ko resolve --filename config | sed "s/VERSION_PLACEHOLDER/${VERSION}/" > "$kf_release"
# append empty config-secrets to kf release
cat << EOF >> "$kf_release"
apiVersion: v1
kind: ConfigMap
metadata:
  annotations:
    operator.knative.dev/mode: Reconcile
  name: config-secrets
  namespace: kf
data:
  wi.googleProjectID: ""
EOF

# copy operator
cp -a operator "${temp_dir}/operator"
cp -a third_party "${temp_dir}/operator/third_party"

# install operator
kf_folder="${temp_dir}/operator/cmd/manager/kodata/kf"
rm -rf "${kf_folder}"/*.yaml
mv "$kf_release" "${kf_folder}/kf.yaml"
pushd "${temp_dir}/operator"
  operator_yaml="${temp_dir}/operator.yaml"
  export KO_DEFAULTBASEIMAGE="gcr.io/distroless/static:nonroot"
  ko resolve --filename config > "$operator_yaml"
  apply_yaml "$operator_yaml"
popd

# TODO: Install Kf system
