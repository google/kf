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


def execute(command):
    call = subprocess.run(command.split(), stdout=subprocess.PIPE, check=True)
    return call.stdout.decode("utf-8")


def list_compute_instances(project_id):
    """Lists the compute instances in the given project."""
    return json.loads(execute("gcloud --project {} compute instances list --format='json'".format(project_id)))


def list_compue_instance_tags(project_id):
    """Returns a set of compute instance tags that are applied to any
    instance.
    """
    tags = set([])

    for instance in list_compute_instances(project_id):
        if 'tags' in instance:
            instance_tags = instance['tags']
            if 'items' in instance_tags:
                tags.update(instance_tags['items'])

    return tags


def list_compute_firewall_rules(project_id):
    """Returns a list of all firewall rules."""
    return json.loads(execute("gcloud --project {} compute firewall-rules list --format='json'".format(project_id)))


def list_abandoned_firewall_rules(project_id):
    """Returns a list of all firewall rules that don't target any known GCE
    instance.
    """
    tags = list_compue_instance_tags(project_id)
    firewall_rules = list_compute_firewall_rules(project_id)

    abandoned = []
    for rule in firewall_rules:
        # don't catch anything without tags
        if 'targetTags' in rule:
            target_tags = set(rule['targetTags'])
            if len(target_tags) == 0:
                # just in case there is an empty targetTags, which is against
                # the API, but they could change in the future.
                print("rule {} targets all resources".format(rule['name']))
            elif tags.isdisjoint(target_tags):
                abandoned.append(rule)
            else:
                print("rule {} is in-use".format(rule['name']))
        else:
            print("rule {} targets all resources".format(rule['name']))

    return abandoned


def delete_abandoned_firewall_rules(project_id, dry_run=False):
    """Deletes each of the abandoned firewall rules returned by
    list_abandoned_firewall_rules.
    """
    abandoned = list_abandoned_firewall_rules(project_id)
    for rule in abandoned:
        name = rule['name']
        print('delete firewall-rule {}'.format(name))
        if dry_run:
            continue
        print(execute('gcloud --quiet --project {} compute firewall-rules delete {}'.format(project_id, name)))


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description='Delete abandoned firewall rules')
    parser.add_argument('project_id', metavar="PROJECT_ID", type=str)
    parser.add_argument('--dry-run', action='store_true')

    args = parser.parse_args()

    delete_abandoned_firewall_rules(args.project_id, dry_run=args.dry_run)
