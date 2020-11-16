#!/usr/bin/env python3

# Copyright 2020 Google LLC
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

import json
import os
import subprocess
import sys


def build_template(dashboard_name, cluster_name, space):
    # Find the template directory. We assume it is in the same location as the
    # script. We assume it is in the same location as the script. We assume it
    # is in the same location as the script. We assume it is in the same
    # location as the script.
    dashboard_path = os.path.join(os.path.dirname(os.path.abspath(__file__)), "dashboard-template.json")

    # Replace the sentinel values.
    # NOTE: Each of these values (e.g., dashboard_name) are required to be
    # valid DNS names and therefore won't have any speical charaters and
    # therefore will not need any escaping.
    with open(dashboard_path) as f:
        return f \
            .read() \
            .replace("XXX-DASHBOARD-XXX", dashboard_name) \
            .replace("XXX-CLUSTER-XXX", cluster_name) \
            .replace("XXX-SPACE-XXX", space)


def execute(command):
    call = subprocess.run(command, stdout=subprocess.PIPE, input=b'y', check=True)
    return call.stdout.decode("utf-8")


def extract_dashboard_id(dashboard_object):
    # example dashboard name: projects/someprojectid/dashboards/dashboard-id
    return dashboard_object["name"].split('/')[3]


def create_dashboard_url(dashboard_id):
    return "https://console.cloud.google.com/monitoring/dashboards/custom/" + dashboard_id


def create_dashboard(dashboard_name, cluster_name, space):
    config = build_template(dashboard_name, cluster_name, space)
    output = execute(["gcloud", "monitoring", "dashboards", "create", "--format", "json", "--config", config])
    return create_dashboard_url(extract_dashboard_id(json.loads(output)))


def main():
    # Ensure we have enough args
    if len(sys.argv) != 4:
        proc_name = sys.argv[0]
        print("Usage: {proc_name} [DASHBOARD NAME] [CLUSTER NAME] [SPACE]")
        sys.exit(1)

    dashboard_name  = sys.argv[1]
    cluster_name  = sys.argv[2]
    space  = sys.argv[3]

    print(create_dashboard(dashboard_name, cluster_name, space))


if __name__ == "__main__":
    main()
