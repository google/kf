---
title: App runtime
description: "Reference guide for the application runtime container environment."
---

The app runtime is the environment apps are executed in.

|                     | Buildpack Apps                     | Container Image Apps              |
| ---                 | ---                                | ---                               |
| System libraries    | Provided by the Stack              | Provided in the container         |
| Network access      | Full access through Envoy sidecar  | Full access through Envoy sidecar |
| File system         | Ephemeral storage                  | Ephemeral storage                 |
| Language runtime    | Supplied by the Stack or Buildpack | Built into the container          |
| User                | Specified by the Stack             | Specified on the container        |
| Isolation mechanism | Kubernetes Pod                     | Kubernetes Pod                    |
| DNS                 | Provided by Kubernetes             | Provided by Kubernetes            |

## Environment variables

Environment variables are injected into the app at runtime by Kubernetes.
Variables are added based on the following order, where later values override
earlier ones with the same name:

1. Space (set by administrators)
1. App (set by developers)
1. System (set by Kf)

Kf provides the following system environment variables:

| Variable                  | Purpose |
| ---                       | ---     |
| `CF_INSTANCE_ADDR`        | The cluster-visible IP:PORT of the App instance. |
| `CF_INSTANCE_GUID`        | The UUID of the App instance. |
| `INSTANCE_GUID`           | Alias of `CF_INSTANCE_GUID`. |
| `CF_INSTANCE_INDEX`       | The index number of the App instance, this will ALWAYS be 0. |
| `INSTANCE_INDEX`          | Alias of `CF_INSTANCE_INDEX`. |
| `CF_INSTANCE_IP`          | The cluster-visible IP of the App instance. |
| `CF_INSTANCE_INTERNAL_IP` | Alias of `CF_INSTANCE_IP` |
| `VCAP_APP_HOST`           | Alias of `CF_INSTANCE_IP` |
| `CF_INSTANCE_PORT`        | The cluster-visible port of the App instance. In Kf this is the same as `PORT`. |
| `DATABASE_URL`            | The first URI found in a `VCAP_SERVICES` credential. |
| `LANG`                    | Required by Buildpacks to ensure consistent script load order. |
| `MEMORY_LIMIT`            | The maximum amount of memory in MB the App can consume. |
| `PORT`                    | The port the App should listen on for requests. |
| `VCAP_APP_PORT`           | Alias of `PORT`. |
| `VCAP_APPLICATION`        | A JSON structure containing App metadata. |
| `VCAP_SERVICES`           | A JSON structure specifying bound services. |


Service credentials from bound services get injected into Apps via the
`VCAP_SERVICES` environment variable. The variable is a valid JSON object with
the following structure.

### VCAPServices

A JSON object where the keys are Service labels and the values are an array of
`VCAPService`. The array represents every bound
service with that label.
[User provided services]({{<relref "user-provided-services">}})
are placed under the label `user-provided`.

**Example**

```.json
{
  "mysql": [...],
  "postgresql": [...],
  "user-provided": [...]
}
```

### VCAPService

This type represents a single bound service instance.

**Example**

```.json
{
  "binding_name": string,
  "instance_name": string,
  "name": string,
  "label": string,
  "tags": string[],
  "plan": string,
  "credentials": object
}
```

**Fields**

| Field           | Type       | Description |
| ---             | ---        | ---         |
| `binding_name`  | `string`   | The name assigned to the service binding by the user. |
| `instance_name` | `string`   | The name assigned to the service instance by the user. |
| `name`          | `string`   | The `binding_name` if it exists; otherwise the `instance_name`. |
| `label`         | `string`   | The name of the service offering. |
| `tags`          | `string[]` | An array of strings an app can use to identify a service instance. |
| `plan`          | `string[]` | The service plan selected when the service instance was created. |
| `credentials`   | `object`   | The service-specific credentials needed to access the service instance. |

{{< note >}} Use the `tags` field to discover services in your App rather than filtering by
plan/label so you can swap implementations and providers per environment.{{< /note >}}

{{< note >}} The `credentials` field is _often_ an object with string keys and values.
However, the values are allowed to be [any type](https://github.com/openservicebrokerapi/servicebroker/blob/v2.15/spec.md#body-9).{{< /note >}}


### VCAP_APPLICATION

The`VCAP_APPLICATION` environment variable is a JSON object containing metadata about the App.

**Example**

```.json
{
  "application_id": "12345",
  "application_name": "my-app",
  "application_uris": ["my-app.example.com"],
  "limits": {
    "disk": 1024,
    "mem": 256
  },
  "name": "my-app",
  "process_id": "12345",
  "process_type": "web",
  "space_name": "my-ns",
  "uris": ["my-app.example.com"]
}
```

**Fields**

| Field                | Type       | Description |
| ---                  | ---        | ---         |
| `application_id`     | `string`   | The GUID identifying the App.               |
| `application_name`   | `string`   | The name assigned to the App when it was pushed. |
| `application_uris`   | `string[]` | The URIs assigned to the App. |
| `limits`             | `object`   | The limits to disk space, and memory permitted to the App. Memory and disk space limits are supplied when the App is deployed, either on the command line or in the App manifest. Disk and memory limits are represented as integers, with an assumed unit of MB. |
| `name`               | `string`   | Identical to `application_name`. |
| `process_id`         | `string`   | The UID identifying the process. Only present in running App containers. |
| `process_type`       | `string`   | The type of process. Only present in running App containers. |
| `space_name`         | `string`   | The human-readable name of the Space where the App is deployed. |
| `uris`               | `string[]` | Identical to `application_uris`.  |

**Missing Fields**

Some fields in `VCAP_APPLICATION` that are in Cloud Foundry are currently not supported in Kf.

Besides CF-specific and deprecated fields (`cf_api`, `host`, `users`) the fields that are not supported in Kf are:

- `application_version` (identical to `version`)
- `organization_id`
- `organization_name`
- `space_id`
- `start` (identical to `started_at`)
- `started_at_timestamp` (identical to `state_timestamp`)
