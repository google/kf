#!/usr/bin/env python3

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

# Consumes the output of ./hack/scrape-supported-gke-versions.py
# and inverts the relationship between Cloud Run and GKE.

import json
import sys

versions = json.loads(sys.stdin.read())
flattened = []
for version in versions:
    for gke_version in version["gke_versions"]:
        flattened.append({
            "gke_version": gke_version,
            "cloud_run_version": version["cloud_run_version"],
        })
flattened.sort(key=lambda x: x["gke_version"])
print(json.dumps(flattened))
