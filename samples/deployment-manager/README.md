Deploying Kf
============

Deploying Kf via [Deployment Manager][deployment-manager] is a great way to
get up and going fast.

## Prerequisites

* [Install gcloud][install-gcloud] and target project
* [Grant Deployment Manager permission to set IAM policies][grant-dm-iam]
  > NOTE: This is necessary to allow Deployment Manager to grant roles to the
  > created Service Account.

## Creating GKE Cluster

### Choose your settings

```sh
CLUSTER_NAME=<cluster name>
```

> NOTE: Be sure to replace the variable with your desired value.

```sh
gcloud deployment-manager \
  deployments create $CLUSTER_NAME \
  --template cluster.py
```

[install-gcloud]:       https://cloud.google.com/sdk/
[grant-dm-iam]:        https://cloud.google.com/deployment-manager/docs/configuration/set-access-control-resources#granting_permission_to_set_iam_policies
[deployment-manager]:  https://cloud.google.com/deployment-manager/
