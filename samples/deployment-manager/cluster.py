# Copyright 2019 Google Inc. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


def generate_config(context):
    name_prefix = context.env['deployment']
    project_id = context.env['project']
    cluster_name = name_prefix
    sa_name = name_prefix + '-sa'
    sa_email = "%s@%s.iam.gserviceaccount.com" % (sa_name, project_id)

    resources = [
        {
            'name': sa_name,
            'type': 'gcp-types/iam-v1:projects.serviceAccounts',
            'properties': {
                'accountId': sa_name,
                'displayName': 'Kf Cluster %s' % cluster_name,
            },
        },
    ]

    roles = [
        # To access GCR
        'roles/storage.admin',
        # Necessary for Stackdriver logging
        'roles/logging.logWriter',
        # Necessary for Stackdriver metrics
        'roles/monitoring.metricWriter',
        # Necessary for controller to be able to manipulate GSAs for Spaces
        'roles/iam.serviceAccountAdmin',
    ]

    resources.extend([
        {
            'name': 'get-bind-role',
            'action': 'gcp-types/cloudresourcemanager-v1:cloudresourcemanager.projects.getIamPolicy',
            'properties': {
                'resource': project_id,
                'options': {
                    'requestedPolicyVersion': 3
                }
            },
        }
    ])

    resources.append(
        {
            'name': 'add-bind-roles',
            'action': 'gcp-types/cloudresourcemanager-v1:cloudresourcemanager.projects.setIamPolicy',
            'properties': {
                'resource': project_id,
                'policy': '$(ref.get-bind-role)',
                'gcpIamPolicyPatch': {
                    'add': [
                        {
                            'role': role,
                            'members': [
                                'serviceAccount:' + sa_email
                            ],
                        } for role in roles
                    ]
                }
            },
            'metadata': {
                'dependsOn': [sa_name],
            },
        }
    )

    auto_scaling = {
        'enabled': context.properties['clusterAutoScaling']
    }

    if context.properties['clusterAutoScaling']:
        auto_scaling['minNodeCount'] = context.properties['clusterAutoScalingMinNodeCount']
        auto_scaling['maxNodeCount'] = context.properties['clusterAutoScalingMaxNodeCount']
        auto_scaling['autoprovisioned'] = False

    resources.append({
        'name': cluster_name,
        'type': 'gcp-types/container-v1beta1:projects.locations.clusters',
        'properties': {
            'parent': 'projects/%s/locations/%s' % (project_id, context.properties['zone']),
            'zone': context.properties['zone'],
            'cluster': {
                'name': cluster_name,
                'addonsConfig': {
                    'httpLoadBalancing': {
                        'disabled': False
                    },
                    'horizontalPodAutoscaling': {
                        'disabled': False
                    },
                    'networkPolicyConfig': {
                        'disabled': False
                    }
                },
                'maintenancePolicy': {
                    'window': {
                        'dailyMaintenanceWindow': {
                            'startTime': '08:00'
                        }
                    }
                },
                'network': context.properties['network'],
                'loggingService': 'logging.googleapis.com/kubernetes',
                'monitoringService': 'monitoring.googleapis.com/kubernetes',
                'ipAllocationPolicy': {
                    'useIpAliases': True
                },
                'networkPolicy': {
                    'provider': "CALICO",
                    'enabled': True
                },
                'nodePools': [{
                    'name': cluster_name,
                    'initialNodeCount': context.properties['initialNodeCount'],
                    'management': {
                        'autoUpgrade': True,
                        'autoRepair': True
                    },
                    'autoscaling': auto_scaling,
                    'config': {
                        'machineType': context.properties['machineType'],
                        'imageType': context.properties['imageType'],
                        'diskType': context.properties['diskType'],
                        'diskSizeGb': context.properties['diskSizeGb'],
                        'metadata': {
                            'disable-legacy-endpoints': 'true'
                        },
                        'serviceAccount': sa_email,
                        'oauthScopes': [
                            'https://www.googleapis.com/auth/' + s
                            for s in [
                                'compute',
                                'devstorage.read_only',
                                'logging.write',
                                'monitoring'
                            ]
                        ]
                    }
                }],
                'releaseChannel': {
                    'channel': context.properties['releaseChannel']
                },
                'workloadIdentityConfig': {
                    'workloadPool': project_id + '.svc.id.goog',
                }
            }
        }
    })

    return {'resources': resources}


if __name__ == "__main__":
    # This is here for testing purposes only. This is not invoked by DM.
    class Context:
        def __init__(self):
            self.env = {
                'deployment': 'some-deployment',
                'project': 'some-project',
            }
            self.properties = {
                'zone': 'some-zone',
                'network': 'some-network',
                'initialNodeCount': 'some-initial-node-count',
                'machineType': 'some-machine-type',
                'imageType': 'some-image-type',
                'diskType': 'some-disk-type',
                'diskSizeGb': 'some-disk-size-gb',
            }

    import json
    print(json.dumps(generate_config(Context()), indent=2))
