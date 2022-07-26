# Copyright 2020 Google LLC
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

import sys
import subprocess
import json


class Cluster:
    def __init__(self, name, region):
        self.name = name
        self.region = region


def execute(command):
    call = subprocess.run(command.split(), stdout=subprocess.PIPE, check=True)
    return call.stdout.decode("utf-8")


def gke_clusters(project_id):
    clusters = json.loads(execute("gcloud container clusters list --project %s --format=json" % (project_id)))
    for cluster in clusters:
        yield Cluster(cluster["name"], cluster["location"])


def delete_gke_cluster(project_id, cluster):
    print("Deleting GKE cluster %s in region %s..." % (cluster.name, cluster.region))
    execute("gcloud --quiet container clusters delete %s --project %s --region %s" % (cluster.name, project_id, cluster.region))


def main():
    if len(sys.argv) != 2:
        print("Usage: %s [PROJECT_ID]" % sys.argv[0])
        sys.exit(1)
    project_id = sys.argv[1]

    for cluster in gke_clusters(project_id):
        delete_gke_cluster(project_id, cluster)


if __name__ == '__main__':
    main()
