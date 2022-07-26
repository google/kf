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

import subprocess
import json
import argparse
import sys


def execute(command):
    call = subprocess.run(command.split(), stdout=subprocess.PIPE, check=True)
    return call.stdout.decode("utf-8")


def delete_repo(project_id, full_repo_name):
    name, location = extract_repo_meta(full_repo_name)
    print(f"deleting {name} (location={location})...")
    print(execute(f"gcloud -q --project {project_id} beta artifacts repositories delete --location {location} {name}"))
    print(f"done deleting {name} (location={location}).")


def extract_repo_meta(full_repo_name):
    # Example full name:
    # projects/kf-int/locations/us-west2/repositories/integration-344537823
    #
    # location: us-west2
    # name: integration-344537823
    parts = full_repo_name.split("/")
    location = parts[3]
    name = parts[5]
    return (name, location)


def delete_repos(project_id):
    repos = json.loads(execute(f"gcloud -q --project {project_id} artifacts repositories list --format=json"))
    for repo in repos:
        if repo.get("format") == "DOCKER":
            delete_repo(project_id, repo.get("name"))


def main():
    if len(sys.argv) != 2:
        print("Usage: %s [PROJECT_ID]" % sys.argv[0])
        sys.exit(1)
    project_id = sys.argv[1]

    delete_repos(project_id)


if __name__ == '__main__':
    main()
