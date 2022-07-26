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
import argparse
import asyncio


def execute(command):
    call = subprocess.run(command.split(), stdout=subprocess.PIPE, check=True)
    return call.stdout.decode("utf-8")


def async_execute(command):
    return asyncio.create_subprocess_shell(
        command, stdout=asyncio.subprocess.PIPE, stderr=asyncio.subprocess.PIPE
    )


class Disk:
    def __init__(self, name, zone):
        self.name = name
        self.zone = zone


def disks(project_id):
    disks = json.loads(execute(f"gcloud --project {project_id} compute disks list --format=json"))
    for disk in disks:
        if "name" in disk and "zone" in disk:
            yield Disk(disk["name"], disk["zone"])


def filter_without_users(project_id, disks):
    for disk in disks:
        d = json.loads(execute(f"gcloud --project {project_id} compute disks describe {disk.name} --format=json --zone {disk.zone}"))
        if "users" not in d or len(d["users"]) == 0:
            yield disk


# We should limit how many of these we're doing at once.
sem = asyncio.Semaphore(10)


async def delete_disk(project_id, disk):
    async with sem:
        print(f"deleting {disk.name} from project {project_id}...")
        p = await async_execute(f"gcloud --quiet --project {project_id} compute disks delete {disk.name} --zone {disk.zone}")
        await p.communicate()
        print(f"done deleting {disk.name} from project {project_id}.")
        return p


def delete_unused_disks(project_id):
    for d in filter_without_users(project_id, disks(project_id)):
        yield delete_disk(project_id, d)


async def main():
    parser = argparse.ArgumentParser(description='Delete abandoned disks')
    parser.add_argument('project_id', metavar="PROJECT_ID", type=str)
    args = parser.parse_args()

    await asyncio.gather(*list(delete_unused_disks(args.project_id)))

if __name__ == '__main__':
    loop = asyncio.get_event_loop()
    try:
        loop.run_until_complete(main())
    finally:
        loop.run_until_complete(loop.shutdown_asyncgens())
        loop.close()
    asyncio.run(main())
