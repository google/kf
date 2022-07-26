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


class Binding:
    def __init__(self, role, member):
        self.role = role
        self.member = member


def execute(command):
    call = subprocess.run(command.split(), stdout=subprocess.PIPE, check=True)
    return call.stdout.decode("utf-8")

def deleted_service_account_bindings(project_id):
    policies = json.loads(execute("gcloud projects get-iam-policy %s --format=json" % (project_id)))
    for binding in policies.get("bindings", []):
        for member in binding.get("members", []):
            if member.startswith("deleted:serviceAccount"):
                yield Binding(binding["role"], member)

def delete_binding(project_id, binding):
    print("Deleting IAM policy binding for %s in role %s..." % (binding.member, binding.role))
    execute("gcloud projects remove-iam-policy-binding %s --role %s --member %s" % (project_id, binding.role, binding.member))

def main():
    if len(sys.argv) != 2:
        print("Usage: %s [PROJECT_ID]" % sys.argv[0])
        sys.exit(1)
    project_id = sys.argv[1]

    for binding in deleted_service_account_bindings(project_id):
        delete_binding(project_id, binding)


if __name__ == '__main__':
    main()

