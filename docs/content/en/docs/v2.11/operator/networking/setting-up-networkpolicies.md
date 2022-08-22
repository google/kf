---
title: Set up network policies
description: Learn to set up a Kubernetes NetworkPolicies to configure traffic.
---

Kf integrates tightly with Kubernetes and Istio to
provide robust network policy enforcement.

By default, Kf workloads are run in the Kubernetes cluster and resolve addresses
using [Kubernetes DNS](https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/).
This DNS resolver will first attempt to resolve addresses within the cluster,
and only if none are found will attempt external resolution.

Each Kf App gets run with an Envoy sidecar injected by [Istio](https://istio.io/)
or the [Anthos Service Mesh](https://cloud.google.com/anthos/service-mesh) (ASM).
This sidecar proxies all network traffic in and out of the Kubernetes Pod.

Each Kubernetes Pod is executed on a Node, a physical or virtual machine
responsible for managing the container images that make up a Pod. Nodes exist on
a physical or virtual network.

Together, these form a hierarchy of systems you can apply network policies.
These are listed below from least to most granular.

## Network level policies

Workload protection starts with the network your GKE cluster is installed on.

If you're running Kf on a GKE cluster on GCP, Kf recommends:

* Placing your GKE cluster on a [Virtual Private Cloud](https://cloud.google.com/vpc/docs) (VPC) network.
  * With [Private Google Access](https://cloud.google.com/vpc/docs/configure-private-google-access) enabled.
* Using [Cloud NAT](https://cloud.google.com/nat) to control egress.

## Node level policies

You can set up policies for containers running on the Node using Kubernetes NetworkPolicies.
These are the closest mapping to Cloud Foundry network policies that exist in Kubernetes.

NetworkPolicies are backed by a Kubernetes add-on. If you set up your own GKE
cluster, you will need to [enable NetworkPolicy enforcement](https://cloud.google.com/kubernetes-engine/docs/how-to/network-policy).

{{< note >}} NetworkPolicies applied by [Project Calico](https://www.projectcalico.org/)
filter on a per-connection basis. On Linux they are backed by [iptables](https://en.wikipedia.org/wiki/Iptables).
Like any firewall, **they are not a substitute of authentication and authorization**.
{{< /note >}}

Kf labels Apps with `kf.dev/networkpolicy=app` and builds with `kf.dev/networkpolicy=build`.
This allows you to target NetworkPolicies directly at Pods running Apps or Builds.

Each Kf Space creates two NetworkPolicies to start with, one targeting Apps and
one targeting Builds. You can change the configuration on the Space's
`spec.networkConfig.(app|build)NetworkPolicy.(in|e)gress` fields.
These fields can be set to one of the following values:

| Enum Value  | Description         |
| ---         | ---                 |
| `PermitAll` | Allows all traffic. |
| `DenyAll`   | Denies all traffic. |

By default Kf uses a permissive network policy. This allows the following
functionality that Kf uses:

* North/South routing to the cluster ingress gateway
* Egress to the Internet e.g. to fetch Buildpacks
* East/West routing between Apps
* Access to the Kubernetes DNS server
* Access to container registries
* Direct access to the VPC network
* Access to Google services like Cloud Logging
* Access to the Workload Identity server for automatic rotating credentials

{{< note >}} Setting a default `DenyAll` policy will break existing and new Apps and Builds
unless you create additional NetworkPolicies to add back the connections listed
above.
{{< /note >}}

## Service mesh policies

If you need fine-grained networking control, authentication, authorization, and
observability you can apply policies using [Anthos Service Mesh](https://cloud.google.com/service-mesh/docs/overview).

A service mesh is an infrastructure layer that enables managed, observable and
secure communication across your services, letting you create robust enterprise
applications made up of many microservices on your chosen infrastructure.

You can see the [list of supported features here](https://cloud.google.com/service-mesh/docs/supported-features).
