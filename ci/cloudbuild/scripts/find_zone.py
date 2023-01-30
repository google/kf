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

def zone_to_region(zone):
    # example: us-central1-a -> us-central1
    return zone[:zone.rfind("-")]

def get_free_quota(region):
    result = json.loads(execute(f"gcloud compute regions describe {region} --format=json"))
    quotas = result["quotas"]
    quotas = {q["metric"]: q for q in quotas}

    # needs to have at least 4 available addresses
    if quotas["IN_USE_ADDRESSES"]["limit"] - quotas["IN_USE_ADDRESSES"]["usage"] < 4:
      return 0
    
    return quotas["CPUS"]["limit"] - quotas["CPUS"]["usage"]

def main():
    # Ensure we have enough args
    if len(sys.argv) != 2:
        proc_name = sys.argv[0]
        print(f"Usage: {proc_name} [MACHINE-TYPE]")
        sys.exit(1)

    machine_type = sys.argv[1]

    zones = list(zones_for_machine(machine_type))
    regions = set(zone_to_region(zone) for zone in zones)

    regions_ordered = sorted(((get_free_quota(region), region) for region in regions))
    if len(regions) == 0:
      print("Failed to find zone with quota")
      sys.exit(2)

    _, region = regions_ordered[-1]
    zone = random.choice([zone for zone in zones if region in zone])
    print(zone)

if __name__ == "__main__":
    main()
