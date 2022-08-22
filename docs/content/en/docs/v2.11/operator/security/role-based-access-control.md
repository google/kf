---
title: "Role-based access control"
description: Learn about how to share a Kf cluster using roles.
weight: 200
---

Kf provides a set of Kubernetes roles that allow multiple teams to share a Kf cluster.
This page describes the roles and best practices to follow when using them.

## When to use Kf roles

Kf roles allow multiple teams to share a Kubernetes
cluster with Kf installed. The roles provide access to
individual Kf Spaces.

Use Kf roles to share access to a cluster if the
following are true:

* The cluster is used by trusted teams.
* Workloads on the cluster share the same assumptions about the level of
  security provided by the environment.
* The cluster exists in a Google Cloud project that is tightly controlled.

Kf roles will not:

* Protect your cluster from untrusted developers or workloads. See the
  [GKE shared responsibility model](https://cloud.google.com/blog/products/containers-kubernetes/exploring-container-security-the-shared-responsibility-model-in-gke-container-security-shared-responsibility-model-gke)
  for more information.
* Provide isolation for your workloads. See the
  [guide to harden your cluster](https://cloud.google.com/kubernetes-engine/docs/how-to/hardening-your-cluster)
  for more information.
* Prevent
  [additional Kubernetes roles](https://cloud.google.com/kubernetes-engine/docs/how-to/role-based-access-control)
  from being defined that interact with Kf.
* Prevent access from [administrators who have access to the Google Cloud project or cluster](https://cloud.google.com/kubernetes-engine/docs/how-to/iam).

## Kf roles {#roles}

The following sections describe the Kubernetes RBAC Roles provided by
Kf and how they interact with GKE IAM.

### Predefined roles {#predefined_roles}

Kf provides several predefined Kubernetes roles to help
you provide access to different subjects that perform different roles. Each
predefined role can be bound to a subject within a Kubernetes Namespace managed
by a Kf Space.

When a subject is bound to a role within a Kubernetes Namespace, their access is
limited to objects that exist in the Namespace that match the grants listed in
the role. In Kf, some resources are defined at the cluster
scope. Kf watches for changes to subjects in the Namespace
and grants additional, limited, roles at the cluster scope.

| Role | Title | Description | Scope |
| ---  | ---   | ---         | ---   |
| `space-auditor` | Space Auditor | Allows read-only access to a Space. | Space |
| `space-developer` | Space Developer | Allows application developers to deploy and manage applications within a Space. | Space |
| `space-manager` | Space Manager | Allows administration and the ability to manage auditors, developers, and managers in a Space. | Space |
| `SPACE_NAME-manager` | Dynamic Space Manager | Provides write access to a single Space object, automatically granted to all subjects with the `space-manager` role within the named Space. | Cluster |
| `kf-cluster-reader` | Cluster Reader | Allows read-only access to cluster-scoped Kf objects, automatically granted to all `space-auditor`, `space-developer`, and `space-manager`. | Cluster |

Information about the policy rules that make up each predefined role can be found in the
[Kf roles reference documentation]({{< relref "kubernetes-roles" >}}).

### Google Cloud IAM roles {#interaction_with_iam}

Kf roles provide access control for objects within a
Kubernetes cluster. Subjects must also be granted an Cloud IAM role to
authenticate to the cluster:

* **Platform administrators** should be granted the `roles/container.admin` Cloud IAM role.
  This will allow them to install, upgrade, and delete Kf as well as create, and delete
  cluster scoped Kf objects like [Spaces or ClusterServiceBrokers](kf-dependencies#custom-resources).

* **Kf end-users** should be granted the `roles/container.viewer` Cloud IAM role.
  This role will allow them to authenticate to a cluster with limited permissions that can be expanded using
  Kf roles.

Google Cloud IAM offers additional [predefined Roles](https://cloud.google.com/iam/docs/understanding-roles#predefined_roles)
for GKE to solve more advanced use cases.

## Map Cloud Foundry roles to Kf

[Cloud Foundry provides roles](https://docs.cloudfoundry.org/concepts/roles.html)
are similar to Kf's [predefined roles](#predefined_roles).
Cloud Foundry has two major types of roles:

* Roles assigned by the User Account and Authentication (UAA) subsystem that provide
  coarse-grained OAuth scopes applicable to all Cloud Foundry API endpoints.
* Roles granted within the Cloud Controller API (CAPI) that provide fine-grained
  access to API resources.

### UAA roles

Roles provided by UAA are most similar to project scoped Google Cloud IAM roles:

* **Admin** users in Cloud Foundry can perform administrative activities for all
  Cloud Foundry organizations and spaces. The role is most similar to the
  `roles/container.admin` Google Cloud IAM role.
* **Admin read-only** users in Cloud Foundry can access all Cloud Foundry API
  endpoints. The role is most similar to the `roles/container.admin`
  Google Cloud IAM role.
* **Global auditor** users in Cloud Foundry have read access to all Cloud Foundry API
  endpoints except for secrets.
  There is no equivalent Google Cloud IAM role, but you can create a
  [custom role](https://cloud.google.com/iam/docs/creating-custom-roles) with similar permissions.

### Cloud Controller API roles

Roles provided by CAPI are most similar to Kf roles granted within a cluster
to subjects that have the `roles/container.viewer` Google Cloud IAM role on the owning project:

* **Space auditors** in Cloud Foundry have read-access to resources in a CF space.
  The role is most similar to the `space-auditor` Kf role.
* **Space developers** in Cloud Foundry have the ability to deploy and manage
  applications in a CF space.
  The role is most similar to the `space-developer` Kf role.
* **Space managers** in Cloud Foundry can modify settings for the CF space and
  assign users to roles.
  The role is most similar to the `space-manager` Kf role.

## What's next

* Learn more about [GKE security in the Security Overview](https://cloud.google.com/kubernetes-engine/docs/concepts/security-overview).
* Make sure you understand the [GKE shared responsibility
  model](https://cloud.google.com/blog/products/containers-kubernetes/exploring-container-security-the-shared-responsibility-model-in-gke-container-security-shared-responsibility-model-gke).
* Learn more about [access control](https://cloud.google.com/kubernetes-engine/docs/concepts/access-control) in GKE.
* Read the [GKE multi-tenancy overview](https://cloud.google.com/kubernetes-engine/docs/concepts/multitenancy-overview).
* Learn about [hardening your GKE cluster](https://cloud.google.com/kubernetes-engine/docs/how-to/hardening-your-cluster).
* Understand the [Kubernetes permissions that make up each Kf predefined role]({{< relref "kubernetes-roles" >}}).
