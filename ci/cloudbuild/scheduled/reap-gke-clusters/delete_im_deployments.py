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
import asyncio


def execute(command):
    call = subprocess.run(command, shell=True, stdout=subprocess.PIPE, input=b'y', check=True)
    return call.stdout.decode("utf-8")


def async_execute(command):
    return asyncio.create_subprocess_shell(
        command, stdout=asyncio.subprocess.PIPE, stderr=asyncio.subprocess.PIPE
    )


def deployments(project_id):
    # List deployments in us-central1 created more than 1 day ago
    cmd = f"gcloud infra-manager deployments list --project {project_id} --location us-central1 --filter='createTime < -P1D' --format=json"
    
    try:
        output = execute(cmd)
        deployments_list = json.loads(output)
    except Exception as e:
        print(f"Error fetching deployments: {e}")
        return

    for deployment in deployments_list:
        # Infra Manager 'name' field is the full resource path:
        # projects/{project}/locations/{location}/deployments/{deployment_id}
        if "name" in deployment:
            yield deployment["name"]


async def delete_deployment(project_id, deployment_full_name):
    print(f"Deleting {deployment_full_name}...")
    
    cmd = f"gcloud infra-manager deployments delete '{deployment_full_name}' --project {project_id} --location us-central1 --quiet"
    
    p = await async_execute(cmd)
    _, stderr = await p.communicate()
    
    if p.returncode == 0:
        print(f"Successfully deleted {deployment_full_name}")
    else:
        print(f"Failed to delete {deployment_full_name}. Error: {stderr.decode('utf-8')}")
    
    return p


def delete_deployments(project_id):
    for deployment_name in deployments(project_id):
        yield delete_deployment(project_id, deployment_name)


async def main():
    if len(sys.argv) != 2:
        print(f"Usage: {sys.argv[0]} [PROJECT-ID]")
        sys.exit(1)
    project_id = sys.argv[1]

    tasks = list(delete_deployments(project_id))
    
    if not tasks:
        print("No old deployments found.")
    else:
        await asyncio.gather(*tasks)


if __name__ == '__main__':
    asyncio.run(main())
