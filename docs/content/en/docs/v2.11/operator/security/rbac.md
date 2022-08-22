---
title: Configure role-based access control
description: Learn to grant users different roles on a cluster.
---

The following steps guide you through configuring role-based access control (RBAC) in a Kf Space.

## Before you begin

Please follow the [GKE RBAC guide](https://cloud.google.com/kubernetes-engine/docs/how-to/role-based-access-control) before continuing with the following steps.

## Configure Identity and Access Management (IAM)

In addition to permissions granted through Kf RBAC, users, groups, or service accounts must also be authenticated to view GKE clusers at the project level. This requirement is the same as for configuring GKE RBAC, meaning users/groups must have at least the `container.clusters.get` IAM permission in the project containing the cluster. This permission is included by the `container.clusterViewer` role, and other more privilleged roles. For more information, review [Interaction with Identity and Access Management](https://cloud.google.com/kubernetes-engine/docs/how-to/role-based-access-control#iam-interaction).

Assign `container.clusterViewer` to a user or group.

```sh
gcloud projects add-iam-policy-binding ${CLUSTER_PROJECT_ID} \
  --role="container.clusterViewer" \
  --member="${MEMBER}"
```

Example member values are:

* user:test-user@gmail.com
* group:admins@example.com
* serviceAccount:test123@example.domain.com

## Manage Space membership as SpaceManager

The cluster admin role, or members with **SpaceManager** role, can assign role to a user, group or service account.

```sh
kf set-space-role MEMBER -t [Group|ServiceAccount|User]
```

The cluster admin role, or members with **SpaceManager** role, can remove a member from a role.

```sh
kf unset-space-role MEMBER -t [Group|ServiceAccount|User]
```

You can view members and their roles within a Space.

```sh
kf space-users
```

### Examples

Assign **SpaceDeveloper** role to a user.

```sh
kf set-space-role alice@example.com SpaceDeveloper
```

Assign **SpaceDeveloper** role to a group.

```sh
kf set-space-role devs@example.com SpaceDeveloper -t Group
```

Assign **SpaceDeveloper** role to a Service Account.

```sh
kf set-space-role sa-dev@example.domain.com SpaceDeveloper -t ServiceAccount
```

## App development as SpaceDeveloper

Members with **SpaceDeveloper** role can perform Kf App development operations within the Space.

To push an App:

```sh
kf push app_name -p [PATH_TO_APP_ROOT_DIRECTORY]
```

To view logs of an App:
```sh
kf logs app_name
```

SSH into a Kubernetes Pod running the App:
```sh
kf ssh app_name
```

View available service brokers:

```sh
kf marketplace
```

## View Apps as SpaceManager or SpaceAuditor

Members with **SpaceManager** or **SpaceAuditor** role could view available Apps within the Space:

```sh
kf apps
```

## View Kf Spaces within a cluster

All roles (**SpaceManager**, **SpaceDeveloper**, and **SpaceAuditor**) can view available Kf Spaces within a cluster:

```sh
kf spaces
```

View Space members and their roles within a Space.

```sh
kf space-users
```

## Impersonation flags

To verify a member's permission, a member with more priviliaged permission can test another member's permissions using the impersonation flags: `--as` and `--as-group`.

For example, as a cluster admin, you can verify if a user (username: bob) has permission to push an App.

```sh
kf push APP_NAME --as bob
```

Verify a group (`manager-group@example.com`) has permission to assign permission to other members.

```sh
kf set-space-role bob SpaceDeveloper --as-group manager-group@example.com
```
