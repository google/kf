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

# This script will returnr the list of zones that work with Artifact Registry
# and that have the correct machine type.

import json
import random
import subprocess
import sys


def execute(command):
    call = subprocess.run(command.split(), stdout=subprocess.PIPE, input=b'y', check=True)
    return call.stdout.decode("utf-8")


def actual_zones():
    zones = json.loads(execute("gcloud compute zones list --format=json"))
    for zone in zones:
        yield zone["name"]


def zones_for_machine(machine_type):
    act_zones = set(actual_zones())
    zones = json.loads(execute(f"gcloud compute machine-types list --format=json"))
    for zone in zones:
        name = zone["name"]
        z = zone["zone"]
        if name == machine_type and z.startswith("us-") and z in act_zones:
            yield z


def regions_for_ar():
    # Only using US regions as other regions are $$$.
    regions = json.loads(execute(f"gcloud artifacts locations list --format=json"))
    for region in regions:
      if region["name"].startswith("us-"):
        yield region["name"]


def zone_to_region(zone):
    # example: us-central1-a -> us-central1
    return zone[:zone.rfind("-")]


QUOTA_REQUIREMENTS = {
    "IN_USE_ADDRESSES": 4,
    "CPUS": 24,
}


def check_quotas(region):
    result = json.loads(execute(f"gcloud compute regions describe {region} --format=json"))
    quotas = result["quotas"]
    quotas = {q["metric"]: q for q in quotas}
    for metric, req in QUOTA_REQUIREMENTS.items():
        quota = quotas[metric]
        if (quota["limit"] - quota["usage"]) < req:
            return False
    return True


def intersection(machine_zones, ar_regions):
    ar_regions = set(ar_regions)
    for zone in machine_zones:
        region = zone_to_region(zone)
        if region in ar_regions:
            yield zone


def main():
    # Ensure we have enough args
    if len(sys.argv) != 2:
        proc_name = sys.argv[0]
        print(f"Usage: {proc_name} [MACHINE-TYPE]")
        sys.exit(1)

    machine_type = sys.argv[1]

    zones = list(intersection(
        zones_for_machine(machine_type),
        filter(check_quotas, regions_for_ar())))
    if len(zones) > 0:
        print(zones[random.randint(0, len(zones)-1)])
    else:
        print("Failed to find zone with quota")
        sys.exit(2)


if __name__ == "__main__":
    main()
