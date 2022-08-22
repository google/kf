---
title: Backing Services Overview
linkTitle: Overview
description: Learn about how Kf works with backing services.
weight: 1
---

Backing services are any processes that the App contacts over the network during its operation.
In traditional operating systems, these services could have been accessed over
the network, a UNIX socket, or could even be a sub-process.
Examples include the following:

* Databases &mdash; for example: MySQL, PostgreSQL
* File storage &mdash; for example: NFS, FTP
* Logging services &mdash; for example: syslog endpoints
* Traditional HTTP APIs &mdash; for example: Google Maps, WikiData, Parcel Tracking APIs

Connecting to backing services over the network rather than installing them all
into the same machine allows developers to focus on their App, independent
security upgrades for different components, and flexibility to swap implementations.

## Backing services in Kf

Kf supports two major types of backing services:

* **Managed services:** These services are created by a service broker and are tied to the Kf cluster.

* **User-provided services:** These services are created outside of Kf, but get can be bound to apps in the same way
as managed services.
