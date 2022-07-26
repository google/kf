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
import http.client
from urllib.parse import urlparse

if len(sys.argv) != 2:
    print("Usage: %s [PROJECT_ID]" % sys.argv[0])
    sys.exit(1)
project_id = sys.argv[1]


def execute(command):
    call = subprocess.run(command.split(), stdout=subprocess.PIPE, check=True)
    return call.stdout.decode("utf-8")


class TargetPool:
    def __init__(self, name, region, health_checks):
        self.name = name
        self.region = region
        self.health_checks = health_checks

    def __hash__(self):
        return hash("%s/%s" % (self.name, self.region))

    def __eq__(self, other):
        return (self.name, self.region) == (other.name, other.region)


def extract_region(regionURL):
    # regionURL paths look like the following:
    # /compute/v1/projects/<PROJECT_ID>/regions/<REGION/6>
    url = urlparse(regionURL)
    splits = url.path.split("/")
    return splits[6]


def target_pools(project_id):
    target_pool_list = json.loads(execute("gcloud --project %s compute target-pools list --format='json'" % project_id))
    for target_pool in target_pool_list:
        health_checks = []
        if "healthChecks" in target_pool:
            health_checks = target_pool["healthChecks"]
        yield TargetPool(target_pool["name"], extract_region(target_pool["region"]), health_checks)


def instances(project_id, target_pool):
    target_pool_desc = json.loads(execute("gcloud --project %s compute target-pools describe --region %s --format='json' %s" % (project_id, target_pool.region, target_pool.name)))
    if "instances" in target_pool_desc:
        for instanceURL in target_pool_desc["instances"]:
            yield instanceURL


def valid_instance(instanceURL):
    url = urlparse(instanceURL)
    conn = http.client.HTTPSConnection(url.netloc)
    auth_token = "Bearer " + execute("gcloud --project %s auth print-access-token" % project_id).strip()
    headers = {"Authorization": auth_token}
    conn.request("GET", url.path, headers=headers)
    return conn.getresponse().getcode() == 200


def valid_target_pool(project_id, target_pool):
    for instanceURL in instances(project_id, target_pool):
        if valid_instance(instanceURL):
            # Found a valid instance, we know the target pool is valid
            return True

    # We didn't find a valid instance, must be invalid
    return False


def map_forwarding_rules(project_id):
    forwarding_rules = json.loads(execute("gcloud --project %s compute forwarding-rules list --format='json'" % project_id))
    result = {}

    # The target pool is under the 'target' field. However it is listed as a
    # URL. The path has the following format:
    # compute/v1/projects/<PROJECT_ID>}/regions/<REGION/6>/targetPools/<TARGET_POOL/8>
    for forwarding_rule in forwarding_rules:
        url = urlparse(forwarding_rule["target"])
        splits = url.path.split("/")
        region = splits[6]
        target_pool_name = splits[8]
        result.update({TargetPool(target_pool_name, region, []): forwarding_rule["name"]})

    return result


def map_health_checks(project_id):
    result = {}
    for target_pool in target_pools(project_id):
        for health_check in target_pool.health_checks:
            result.update({health_check: target_pool.name})

    return result


def health_checks(project_id):
    heath_check_list = json.loads(execute(f"gcloud --project {project_id} compute http-health-checks list --format='json'"))
    for health_check in heath_check_list:
        if "name" in health_check:
            yield health_check["name"]


# We'll cache these so we don't have to do it multiple times.
forwarding_rules = map_forwarding_rules(project_id)
mapped_health_checks = map_health_checks(project_id)


def delete_associated_forwarding_rule(project_id, target_pool):
    if target_pool not in forwarding_rules:
        # Looks like we don't know about an associated forwarding rule
        print("did not find a forwarding rule for target pool %s (region %s)" % (target_pool.name, target_pool.region))
        return

    forwarding_rule_name = forwarding_rules[target_pool]
    print("delete forwarding rule %s (associated with target pool %s)" % (forwarding_rule_name, target_pool.name))
    print(execute("gcloud --quiet --project %s compute forwarding-rules delete --region %s %s" % (project_id, target_pool.region, forwarding_rule_name)))


def delete_target_pool(project_id, target_pool):
    delete_associated_forwarding_rule(project_id, target_pool)
    print(f"deleting target-pool {target_pool.name} in zone {target_pool.region}")
    print(execute("gcloud --quiet --project %s compute target-pools delete --region %s %s" % (project_id, target_pool.region, target_pool.name)))


def delete_health_check(project_id, health_check):
    print(f"deleting HTTP health check {health_check}...")
    print(execute(f"gcloud --quiet --project {project_id} compute http-health-checks delete {health_check}"))


def delete_abandoned_target_pools(project_id):
    for target_pool in target_pools(project_id):
        if valid_target_pool(project_id, target_pool):
            print("target pool %s (region %s) is valid" % (target_pool.name, target_pool.region))
        else:
            print("target pool %s (region %s) is not valid... deleting" % (target_pool.name, target_pool.region))
            delete_target_pool(project_id, target_pool)


# delete_abandoned_health_checks looks for all the HTTP health checks that
# don't have an associated target pool. Any it finds, it deletes.
def delete_abandoned_health_checks(project_id):
    for health_check in health_checks(project_id):
        if health_check not in mapped_health_checks:
            delete_health_check(project_id, health_check)


def main():
    delete_abandoned_target_pools(project_id)
    delete_abandoned_health_checks(project_id)


if __name__ == '__main__':
    main()
