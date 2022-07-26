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

import sys
import subprocess
import json


def execute(command):
    call = subprocess.run(command.split(), stdout=subprocess.PIPE, check=True)
    return call.stdout.decode("utf-8")


def hub_memberships(project_id):
    memberships = json.loads(execute("gcloud container hub memberships list --project %s --format=json" % (project_id)))
    for membership in memberships:
        yield membership["name"].split('/')[-1]

def gke_clusters(project_id):
    clusters = json.loads(execute("gcloud container clusters list --project %s --format=json" % (project_id)))
    return set([cluster["name"] for cluster in clusters])


def delete_membership(project_id, membership):
    execute("gcloud --quiet container hub memberships delete %s --project %s" % (membership, project_id))


def main():
    if len(sys.argv) != 2:
        print("Usage: %s [PROJECT_ID]" % sys.argv[0])
        sys.exit(1)
    project_id = sys.argv[1]

    existing_clusters = gke_clusters(project_id)
    for membership in hub_memberships(project_id):
        if membership not in existing_clusters:
            delete_membership(project_id, membership)


if __name__ == '__main__':
    main()
