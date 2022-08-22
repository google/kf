---
title: Service discovery
weight: 100
description: "Connect to other Apps in the Kf cluster."
---

This document is an overview of Kubernetes DNS-based service discovery
and how it can be used with Kf.

## When to use Kubernetes service discovery with Kf

Kubernetes service discovery can be used by applications that need to locate
backing services in a consistent way regardless of where the application is
deployed. For example, a team might want to use a common URI in their
configuration that always points at the local SMTP gateway to decouple code from
the environment it ran in.

Service discovery helps application teams by:

*   Reducing the amount of per-environment configuration.
*   Decoupling client and server applications.
*   Allowing applications to be portable to new environments.

You can use Kubernetes service discovery when:

*   Applications use their container's DNS configurations to resolve hosts.
*   Applications are deployed with their backing services in the same Kubernetes cluster or namespace.
*   Backing services have an associated
    [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/).
    Kf creates these for each app.
*   Kubernetes NetworkPolicies allow traffic between an application and the
    Kubernetes service it needs to communicate with. Kf
    creates these policies in each Kf space.

You **should not** use Kubernetes service discovery if:

*   Applications need to failover between multiple clusters.
*   You override the DNS resolver used by your application.
*   Applications need specific types of load balancing.

## How Kubernetes service discovery works

Kubernetes service discovery works by
[modifying the DNS configuration](https://cloud.google.com/kubernetes-engine/docs/concepts/service-discovery)
of containers running on a Kubernetes node. When an application looks up an
unqualified domain name, the local DNS resolver will first attempt to resolve
the name in the local cluster.

Domains without multiple parts will be resolved against the names of Kubernetes
services in the container's namespace. Each Kf app
creates a Kubernetes service with the same name. If two
Kf apps `ping` and `pong` were deployed in the same
Kf space, then `ping` could use the URL `http://pong` to
send traffic to the other service.

Domains with a single dot will be resolved against the Kubernetes services in
the Kubernetes namespace with the same name as the label after the dot. For
example, if there was a PostgreSQL database with a `customers` service in the
`database` namespace, an application in another namespace could
resolve it using `postgres://customers.database`.

## How to use service discovery with Kf

Kubernetes DNS based service discovery can be used in any
Kf app. Each Kf app creates a
Kubernetes service of the same name, and each Kf
space creates a Kubernetes namespace with the same name.

*   Refer to a Kf app in the current space using
    <code><var>protocol</var>://<var>app-name</var></code>.
*   Refer to a Kf app in a different space using
    <code><var>protocol</var>://<var>app-name</var>.<var>space-name</var></code>.
*   Refer to a Kf app in the current space listening on
    a custom port using
    <code><var>protocol</var>://<var>app-name</var>:<var>port</var></code>.
*   Refer to a Kf app in a different space listening
    a custom port using
    <code><var>protocol</var>://<var>app-name</var>.<var>space-name</var>:<var>port</var></code>.


{{< note >}} For Kubernetes services managed by Kf, TCP port 80
always maps to the same port in the app's `PORT` environment variable. This is
even true if the Kf app doesn't serve HTTP traffic.{{< /note >}}

## Best practices

Applications that are going to be the target of DNS based service discovery
should have frequent [health checks]({{< relref "manifest#health_check_types" >}})
to ensure they are rapidly added and removed from the pool of hosts that accept
connections.

Applications using DNS based service discovery should not cache the IP addresses
of the resolved services because they are not guaranteed to be stable.

If environment specific services exist outside of the cluster, they can be
resolved using Kubernetes DNS if you set up
[ExternalName Kubernetes services](https://kubernetes.io/docs/concepts/services-networking/service/#externalname).
These Kubernetes services provide the same resolution capabilities, but return a
CNAME record to redirect requests to an external authority.

## Comparison to Eureka

Eureka is an open source client-side load-balancer created by Netflix. It is
commonly used as part of the
[Spring Cloud Services](https://docs.pivotal.io/spring-cloud-services/3-1/common/service-registry/index.html)
service broker. Eureka was
[built to be a regional load balancer and service discovery mechanism](https://github.com/Netflix/eureka/wiki/Eureka-at-a-glance#what-is-eureka)
for services running in an environment that caused frequent disruptions to
workloads leading to unstable IP addresses.

Eureka is designed as a client/server model. Clients register themselves with
the server indicating which names they want to be associated with and
periodically send the server heartbeats. The server allows all connected clients
to resolve names.

In general, you should use Kubernetes DNS rather than Eureka in Kubernetes for
the following reasons:

*   DNS works with all programming languages and applications without the need for libraries.
*   Your application's existing health check will be reused reducing combinations of errors.
*   Kubernetes manages the DNS server, allowing you to rely on fewer dependencies.
*   Kubernetes DNS respects the same policy and RBAC constraints as the rest of Kubernetes.

There are a few times when deploying a Eureka server would be advantageous:

*   You need service discovery across Kubernetes and VM based applications.
*   You need client based load-balancing.
*   You need independent health checks.

## What's next

*   [Read more about service discovery in GKE](https://cloud.google.com/kubernetes-engine/docs/concepts/service-discovery).
*   Learn about [Service Directory](https://cloud.google.com/service-directory), a managed offering similar to Eureka.
