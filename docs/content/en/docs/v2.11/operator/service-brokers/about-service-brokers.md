---
title: "Service Brokers Overview"
linkTitle: Overview
weight: 100
---

Kf supports binding and provisioning apps to [Open Service Broker (OSB)](https://www.openservicebrokerapi.org/) services.

Any compatible service broker can be installed using the `create-service-broker` command, but only the [Kf Cloud Service Broker]({{< relref "cloud-sb-overview" >}}) is fully supported.

{{< note >}} Kf only supports a subset of the types of services
that Open Service Brokers provide. Specifically, it only supports credential
services.{{< /note >}}

Special services such as syslog drains, volume services, route services,
service keys, and shared services aren't currently supported.

