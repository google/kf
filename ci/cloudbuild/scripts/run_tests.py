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
import sys
import asyncio


class BuildStreamer:
    def __init__(self, build_id, process):
        self.build_id = build_id
        self.process = process


def execute(command):
    call = subprocess.run(command.split(), stdout=subprocess.PIPE, input=b'y', check=True)
    return call.stdout.decode("utf-8")


def async_execute(command):
    return asyncio.create_subprocess_shell(
        command, stdout=asyncio.subprocess.PIPE, stderr=asyncio.subprocess.PIPE
    )


def submit_build(release_bucket, node_count, machine_type):
    return execute(f'''
        gcloud builds submit . \
        --config=ci/cloudbuild/test.yaml \
        --machine-type=n1-highcpu-8 \
        --substitutions=_FULL_RELEASE_BUCKET={release_bucket},_DELETE_CLUSTER=true,_NODE_COUNT={node_count},_MACHINE_TYPE={machine_type} \
        --async \
        --format=value(id)
    ''')


def submit_builds(release_bucket, node_counts, machine_types):
    for i in range(len(node_counts)):
        build_id = submit_build(
            release_bucket,
            node_counts[i],
            machine_types[i],
        )
        yield build_id


def build_log_url(build_id):
    return execute(f'gcloud builds describe {build_id} --format=value(logUrl)')


def build_status(build_id):
    return execute(f'gcloud builds describe {build_id} --format=value(status)')


async def stream_build_logs(build_id):
    p = await async_execute(f'gcloud builds log --stream {build_id}')
    await p.communicate()
    return p


def build_log_streamers(build_ids):
    for build_id in build_ids:
        p = stream_build_logs(build_id)
        yield BuildStreamer(build_id, p)


def split_scenarios(scenarios):
    # Depending how the user entered their substitutions, they may have
    # accidently added extra quotes. Just strip them out.
    scenarios = scenarios.strip('"')
    scenarios = scenarios.strip("'")
    return scenarios.split()


async def run(release_bucket, scenario_node_counts, scenario_machine_types):
    # Execute and wait for builds. Failed builds will have their build ID
    # returned.
    build_streamers = build_log_streamers(
        submit_builds(
            release_bucket,
            scenario_node_counts,
            scenario_machine_types
        )
    )

    # Wait for all the builds to finish
    build_streamer_list = list(build_streamers)
    results = await asyncio.gather(*[b.process for b in build_streamer_list])

    # Look at each build and find the failed ones.
    failed = 0
    for i in range(len(results)):
        if results[i].returncode == 0:
            continue

        failed = failed + 1

        # Display the log URL for the failed build.
        build_id = build_streamer_list[i].build_id
        log_url = build_log_url(build_id)
        print(f"Build {build_id} failed: {log_url}")

    node_counts_len = len(scenario_node_counts)
    if failed != 0:
        print(f"{failed} of {node_counts_len} failed")
        sys.exit(1)
    print(f"{node_counts_len} of {node_counts_len} succeeded")


def main():
    # Ensure we have enough args
    if len(sys.argv) != 4:
        proc_name = sys.argv[0]
        print(f"Usage: {proc_name} [RELEASE-BUCKET] [SCENARIO-NODE-COUNTS] [SCENARIO-MACHINE-TYPES]")
        sys.exit(1)

    # Fetch args
    release_bucket = sys.argv[1]
    scenario_node_counts_str = sys.argv[2]
    scenario_machine_types_str = sys.argv[3]

    # Convert to lists
    scenario_node_counts = split_scenarios(scenario_node_counts_str)
    scenario_machine_types = split_scenarios(scenario_machine_types_str)

    # Ensure the scenario lenghts are the same.
    node_counts_len = len(scenario_node_counts)
    machine_types_len = len(scenario_machine_types)
    if len(scenario_node_counts) != len(scenario_machine_types):
        print(f"_SCENARIOS_NODE_COUNTS (len={node_counts_len}) must have the same length as _SCENARIOS_MACHINE_TYPES (len={machine_types_len})")
        sys.exit(1)

    asyncio.run(run(release_bucket, scenario_node_counts, scenario_machine_types))


if __name__ == "__main__":
    main()
