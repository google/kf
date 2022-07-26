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

import subprocess
import json
import sys
import re


def execute(command):
    call = subprocess.run(command.split(), stdout=subprocess.PIPE, input=b'y', check=True)
    return call.stdout.decode("utf-8")


def find_service_accounts(project_id):
    policies = json.loads(execute("gcloud projects get-iam-policy {} --format=json".format(project_id)))
    for binding in policies["bindings"]:
        for member in binding["members"]:
            if re.match('^.+@cloudservices.gserviceaccount.com$', member):
                yield member, "role" in binding and "roles/owner" in binding["role"]


def find_owner(project_id):
    last_service_account = ""
    for service_account, is_owner in find_service_accounts(project_id):
        last_service_account = service_account
        if is_owner:
            return service_account, True
    # Getting here implies we didn't find an owner.
    return last_service_account, False


def main():
    if len(sys.argv) != 2:
        print("Usage: {} [PROJECT_ID]".format(sys.argv[0]))
        sys.exit(1)
    project_id = sys.argv[1]

    service_account, is_owner = find_owner(project_id)

    # Check to see if the service account has the owner role. If not, make it
    # one.
    if not is_owner:
        print("Service Account {} is not an owner. Making it one...".format(service_account))
        execute("gcloud projects add-iam-policy-binding {} --member {} --role roles/owner".format(project_id, service_account))
    else:
        print("Service Account {} is already an owner. Moving on...".format(service_account))


if __name__ == "__main__":
    main()
