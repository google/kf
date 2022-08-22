---
title: Security Overview
linkTitle: Overview
description: Unerstand Kf's security posture.
weight: 100
---

Kf aims to provide a similar developer experience to Cloud Foundry, replicating the build, push, and deploy lifecycle. It does this by building a developer UX layer on top of widely-known, broadly used and adopted technologies like Kubernetes, Istio, and container registries rather than by implementing all the pieces from the ground up.

{{< note >}} Kf should be used in a GCP Project dedicated to your evaluation. See [Important considerations](#important-considerations) for more information.{{< /note >}}

When making security decisions, Kf aims to provide complete solutions that are native to their respective components and can be augmented with other mechanisms. Breaking that down:

* **Complete solutions** means that Kf tries not to provide partial solutions that can lead to a false sense of security.
* **Native** means that the solutions should be a part of the component rather than a Kf construct to prevent breaking changes.
* **Can be augmented** means the approach Kf takes should work seamlessly with other Kubernetes and Google Cloud tooling for defense in depth.

## Important considerations

In addition to the [Current limitations](#current-limitations) described below, it is important that you read through and understand the items outlined in this section.

### Workload Identity

By default, Kf uses [Workload Identity](https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity) to provide secure delivery and rotation of the Service Account credentials used by Kf to interact with your Google Cloud Project. Workload Identity achieves this by mapping a Kubernetes Service Account (KSA) to a Google Service Account (GSA). The Kf controller runs in the `kf` namespace and uses a KSA named `controller` mapped to your GSA to do the following things:

1. Write metrics to Stackdriver
1. When a new Kf space is created (`kf create-space`), the Kf controller creates a new KSA named `kf-builder` in the new space and maps it to the same GSA.
1. The `kf-builder` KSA is used by Tekton to push and pull container images to Google Container Registry (gcr.io)

This diagram illustrates those interactions:


{{<figure src="./wi_overview.svg" alt="Workload identity overview diagram" >}}

### Current limitations

* Kf doesn't provide pre-built RBAC roles. Until
  Kf provides this, use
  [RBAC](https://cloud.google.com/kubernetes-engine/docs/how-to/role-based-access-control).

* A developer pushing an app with Kf can also create
  Pods (with `kubectl`) that can use the `kf-builder` KSA with the permissions
  of its associated GSA.

* Deploying to Kf requires write access to a container
  registry. Deploy Kf in a dedicated project without
  access to production resources. Grant developers access to push code to the
  Artifact Repository by
  [granting them `roles/storage.admin`](https://cloud.google.com/container-registry/docs/access-control)
  on the project or buckets that Artifact Repository uses.

* Kf uses the same Pod to fetch, build, and store images.
  Assume that any credentials that you provide can be known by the authors and
  publishers of the buildpacks you use.

* Kf doesn't support quotas to protect against noisy
  neighbors. Use Kubernetes
  [resource quotas](https://kubernetes.io/docs/concepts/policy/resource-quotas/).

## Other resources

### Google Cloud

#### General

  * [GKE security overview](https://cloud.google.com/kubernetes-engine/docs/concepts/security-overview)
  * [GKE cluster multi-tenancy](https://cloud.google.com/kubernetes-engine/docs/concepts/multitenancy-overview)
  * [Workload Identity](https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity)
  * [GKE and Cloud IAM policies](https://cloud.google.com/kubernetes-engine/docs/how-to/iam)

#### Recommended protections

  * [Protecting cluster metadata](https://cloud.google.com/kubernetes-engine/docs/how-to/protecting-cluster-metadata)
  * [Role-based access control](https://cloud.google.com/kubernetes-engine/docs/how-to/role-based-access-control)

#### Advanced protections

  * [GKE Sandbox](https://cloud.google.com/kubernetes-engine/docs/how-to/sandbox-pods)
  * [Network policies](https://cloud.google.com/kubernetes-engine/docs/how-to/network-policy)
