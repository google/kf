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
from concurrent.futures import ThreadPoolExecutor


def execute(command):
    call = subprocess.run(command.split(), stdout=subprocess.PIPE, check=True)
    return call.stdout.decode("utf-8")


def delete_image(project_id, full_image_name):
    print(f"deleting {full_image_name}...")
    print(execute(f"gcloud -q --project {project_id} container images delete --force-delete-tags {full_image_name}"))
    print(f"done deleting {full_image_name}.")


def attach_digests(project_id, image_name):
    images = json.loads(execute(f"gcloud --project {project_id} container images list-tags --format=json {image_name}"))
    for image in images:
        digest = image.get("digest")
        yield f"{image_name}@{digest}"


def protected_image(project_id, image_name):
    return image_name == f"gcr.io/{project_id}/reap-gke-clusters" or image_name == f"gcr.io/{project_id}/ko"


def delete_images(project_id):
    images = json.loads(execute(f"gcloud --project {project_id} container images list --format=json"))

    with ThreadPoolExecutor(max_workers=16) as executor:
      for image in images:
          image_name = image.get("name")

          if protected_image(project_id, image_name):
              continue

          for full_name in attach_digests(project_id, image_name):
            executor.submit(delete_image, project_id, full_name)


def main():
    if len(sys.argv) != 2:
        print("Usage: %s [PROJECT_ID]" % sys.argv[0])
        sys.exit(1)
    project_id = sys.argv[1]

    delete_images(project_id)


if __name__ == '__main__':
    main()
