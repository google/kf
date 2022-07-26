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
from os import path


def execute(command):
    call = subprocess.run(command, stdout=subprocess.PIPE, check=True)
    return call.stdout.decode("utf-8")


def execute_str(command):
    return execute(command.split())


class Job:
    def __init__(self, name, zone):
        self.name = name
        self.zone = zone


def parse_zone(name):
    # Example name:
    # projects/<PROJECT-ID>/locations/<ZONE>/jobs/<NAME>
    return name.split("/")[3]


def scheduler_jobs(project_id, name):
    jobs = json.loads(execute_str("gcloud scheduler jobs list --project %s --filter name:%s --format json" % (project_id, name)))
    for job in jobs:
        yield Job(name, parse_zone(job["name"]))


def delete_scheduler_job(project_id, job):
    print("deleting scheduler job %s in zone %s..." % (job.name, job.zone))
    execute_str("gcloud --quiet scheduler jobs delete --project %s %s" % (project_id, job.name))


def create_scheduler_job(project_id, job_name, service_account, schedule, message_body):
    # If the job already exists, delete it.
    for job in scheduler_jobs(project_id, job_name):
        delete_scheduler_job(project_id, job)
        break

    print("creating scheduler job %s..." % job_name)

    execute([
        "gcloud", "scheduler", "jobs", "create",
        "http",
        job_name,
        "--quiet",
        "--project", project_id,
        "--schedule", schedule,
        "--time-zone", "America/Los_Angeles",
        "--uri", "https://cloudbuild.googleapis.com/v1/projects/%s/builds" % project_id,
        "--http-method", "POST",
        "--headers", "Content-Type=application/json",
        "--oauth-service-account-email", service_account,
        "--oauth-token-scope", "https://www.googleapis.com/auth/cloud-platform",
        "--message-body", message_body
    ])


def main():
    if len(sys.argv) != 6:
        print("Usage: %s [PROJECT_ID] [NAME] [SERVICE_ACCOUNT] [SCHEDULE] [MESSAGE_BODY_FILEPATH]" % sys.argv[0])
        sys.exit(1)
    project_id = sys.argv[1]
    job_name = sys.argv[2]
    service_account = sys.argv[3]
    schedule = sys.argv[4]
    message_body_filepath = sys.argv[5]

    if message_body_filepath == "-":
        # Use STDIN
        message_body = sys.stdin.read()
    elif path.exists(message_body_filepath):
        # Read from file
        with open(message_body_filepath, "r") as f:
            message_body = f.read()
    else:
        # Assume the passed in argument is data
        message_body = message_body_filepath

    create_scheduler_job(project_id, job_name, service_account, schedule, message_body)


if __name__ == '__main__':
    main()
