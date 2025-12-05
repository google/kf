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

set -euxo pipefail

_GCP_PROJECT_ID="${GCP_PROJECT_ID}"

cd ci/cloudbuild/scheduled/reap-gke-clusters

python3 delete_im_deployments.py ${_GCP_PROJECT_ID}
python3 delete_gke_clusters.py ${_GCP_PROJECT_ID}
python3 cleanup_load_balancers.py ${_GCP_PROJECT_ID}
python3 cleanup_firewall_rules.py ${_GCP_PROJECT_ID}
python3 delete_disks.py ${_GCP_PROJECT_ID}
python3 reap_gcr_containers.py ${_GCP_PROJECT_ID}
python3 reap_ar_repos.py ${_GCP_PROJECT_ID}
python3 reap_iam_bindings.py ${_GCP_PROJECT_ID}
python3 reap_hub_memberships.py ${_GCP_PROJECT_ID}