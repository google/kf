---
title: Kubernetes Roles
description: Understand how Kf uses Kubernetes' RBAC to assign roles.
---
The following sections describe the [Kubernetes ClusterRoles](https://kubernetes.io/docs/reference/access-authn-authz/rbac/)
that are created by Kf and lists the permissions that are
contained in each ClusterRole.

## Space developer role {#space-developer}

The Space developer role aggregates permissions application developers use
to deploy and manage applications within a Kf Space.

You can retrieve the permissions granted to Space developers on your cluster
using the following command.

```sh
kubectl describe clusterrole space-developer
```

The default installation of Kf provides the following
permissions:

```none
PolicyRule:
  Resources                               Non-Resource URLs  Resource Names  Verbs
  ---------                               -----------------  --------------  -----
  events                                  []                 []              [*]
  secrets                                 []                 []              [*]
  *.kf.dev                                []                 []              [*]
  networkpolicies.networking.k8s.io       []                 []              [*]
  pods/exec                               []                 []              [create]
  *.upload.kf.dev                         []                 []              [create]
  pods/log                                []                 []              [get list watch]
  pods                                    []                 []              [get list watch]
  rolebindings.rbac.authorization.k8s.io  []                 []              [get list watch]
```

## Space auditor role {#space-auditor}

The Space auditor role aggregates read-only permissions that auditors and
automated tools use to validate applications within a
Kf Space.

You can retrieve the permissions granted to Space auditors on your cluster
using the following command.

```sh
kubectl describe clusterrole space-auditor
```

The default installation of Kf provides the following
permissions:

```none
PolicyRule:
  Resources                               Non-Resource URLs  Resource Names  Verbs
  ---------                               -----------------  --------------  -----
  events                                  []                 []              [*]
  apps.kf.dev                             []                 []              [get list watch]
  rolebindings.rbac.authorization.k8s.io  []                 []              [get list watch]
```

## Space manager role {#space-manager}

The Space manager role aggregates permissions that allow delegation of duties to
others within a Kf Space.

You can retrieve the permissions granted to Space managers on your cluster
using the following command.

```sh
kubectl describe clusterrole space-manager
```

The default installation of Kf provides the following
permissions:

```none
PolicyRule:
  Resources                               Non-Resource URLs  Resource Names     Verbs
  ---------                               -----------------  --------------     -----
  clusterroles.rbac.authorization.k8s.io  []                 [space-auditor]    [bind]
  clusterroles.rbac.authorization.k8s.io  []                 [space-developer]  [bind]
  clusterroles.rbac.authorization.k8s.io  []                 [space-manager]    [bind]
  rolebindings.rbac.authorization.k8s.io  []                 []                 [get list update patch watch]
  apps.kf.dev                             []                 []                 [get list watch]
```

{{< note >}} Subjects bound to the `space-manager` ClusterRole within a
Kf Space are also granted write access to that Space.
{{< /note >}}

## Dynamic Space manager role {#dynamic-space-manager}

Each Kf Space creates a ClusterRole with the name
`SPACE_NAME-manager`, where
`SPACE_NAME-manager` is called the dynamic manager role.

Kf
automatically grants all subjects with the `space-manager` role within the
Space the dynamic manager role at the cluster scope. The permissions for the
dynamic manager role allow Space managers to update settings on the Space with
the given name.

You can retrieve the permissions granted to the dynamic manager role for any
Space on your cluster using the following command.

```sh
kubectl describe clusterrole SPACE_NAME-manager
```

The default installation of Kf provides the following
permissions:

```none
PolicyRule:
  Resources      Non-Resource URLs  Resource Names  Verbs
  ---------      -----------------  --------------  -----
  spaces.kf.dev  []                 [SPACE_NAME]    [get list watch update patch]
```

## Kf cluster reader role {#kf-cluster-reader}

Kf automatically grants the `kf-cluster-reader` role to all users on a
cluster that already have the `space-developer`, `space-auditor`, or `space-manager`
role within a Space.

You can retrieve the permissions granted to Space Kf
cluster readers on your cluster using the following command.

```sh
kubectl describe clusterrole kf-cluster-reader
```

The default installation of Kf provides the following
permissions:

```none
PolicyRule:
  Resources                     Non-Resource URLs  Resource Names  Verbs
  ---------                     -----------------  --------------  -----
  namespaces                    []                 [kf]            [get list watch]
  clusterservicebrokers.kf.dev  []                 []              [get list watch]
  spaces.kf.dev                 []                 []              [get list watch]
```