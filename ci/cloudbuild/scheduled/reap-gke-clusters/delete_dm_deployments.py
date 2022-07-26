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
    call = subprocess.run(command.split(), stdout=subprocess.PIPE, input=b'y', check=True)
    return call.stdout.decode("utf-8")


def async_execute(command):
    return asyncio.create_subprocess_shell(
        command, stdout=asyncio.subprocess.PIPE, stderr=asyncio.subprocess.PIPE
    )


def deployments(project_id):
    deployments = json.loads(execute(f"gcloud --project {project_id} deployment-manager deployments list --format=json"))
    for deployment in deployments:
        if "name" in deployment:
            yield deployment["name"]


async def delete_deployment(project_id, deployment_name):
    print(f"deleting {deployment_name} from project {project_id}...")
    p = await async_execute(f"gcloud --quiet --project {project_id} deployment-manager deployments delete {deployment_name}")
    await p.communicate()
    print(f"done deleting {deployment_name} from project {project_id}")
    return p


def delete_deployments(project_id):
    for deployment_name in deployments(project_id):
        yield delete_deployment(project_id, deployment_name)


async def main():
    if len(sys.argv) != 2:
        print(f"Usage: {sys.argv[0]} [PROJECT-ID]")
        sys.exit(1)
    project_id = sys.argv[1]

    await asyncio.gather(*list(delete_deployments(project_id)))

if __name__ == '__main__':
    asyncio.run(main())
