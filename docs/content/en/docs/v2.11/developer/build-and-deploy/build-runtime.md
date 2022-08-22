---
title:  Build runtime
description: "Reference guide for the application build container environment."
---

The Build runtime is the environment Apps are built in.

|                     | Buildpack Builds                   | Docker Builds |
| ---                 | ---                                | ---           |
| System libraries    | Provided by the Stack              | User supplied |
| Network access      | Full access through Envoy sidecar  | Full access through Envoy sidecar |
| File system         | No storage                         | No storage |
| Language runtime    | Provided by the Stack              | User supplied |
| User                | Specified by the Stack             | User supplied |
| Isolation mechanism | Kubernetes Pod                     | Kubernetes Pod |
| DNS                 | Provided by Kubernetes             | Provided by Kubernetes |

## Environment variables

Environment variables are injected into the Build at runtime.
Variables are added based on the following order, where later values override
earlier ones with the same name:

1. Space (set by administrators)
1. App (set by developers)
1. System (set by Kf)

Kf provides the following system environment variables to Builds:

| Variable                  | Purpose |
| ---                       | ---     |
| `CF_INSTANCE_ADDR`        | The cluster-visible IP:PORT of the Build. |
| `INSTANCE_GUID`           | Alias of `CF_INSTANCE_GUID`. |
| `CF_INSTANCE_IP`          | The cluster-visible IP of the Build. |
| `CF_INSTANCE_INTERNAL_IP` | Alias of `CF_INSTANCE_IP` |
| `VCAP_APP_HOST`           | Alias of `CF_INSTANCE_IP` |
| `CF_INSTANCE_PORT`        | The cluster-visible port of the Build. |
| `LANG`                    | Required by Buildpacks to ensure consistent script load order. |
| `MEMORY_LIMIT`            | The maximum amount of memory in MB the Build can consume. |
| `VCAP_APPLICATION`        | A JSON structure containing App metadata. |
| `VCAP_SERVICES`           | A JSON structure specifying bound services. |

{{< note >}} The environment variables Kf provides to Builds are a subset of those provided
to [App Runtime]({{<relref "app-runtime">}}).{{< /note >}}

