---
title: "App Manifests"
linkTitle: "App Manifests"
weight: 10
---

App manifests provide a way for developers to record their app's execution environment in a declarative way.
They allow applications to be deployed consistently and reproducibly.

## Format

Manifests are YAML files in the root directory of the app. They MUST be named `manifest.yml` or `manifest.yaml`.

Kf app manifests are allowed to have a single top-level element: `applications`.
The `applications` element can contain one or more application entries.

## Application Fields

The following fields are valid for objects under `applications`:

| Field | Type | Description |
|:------|:-----|:------------|
| **name** | string | The name of the application. The app name should be lower-case alphanumeric characters and dashes. It must not start with a dash. |
| **path** | string | The path to the source of the app. Defaults to the manifest's directory. |
| **buildpacks** | string[] | A list of buildpacks to apply to the app. |
| **stack** | string | Base image to use for to use for apps created with a buildpack. |
| **docker** | object | A docker object. See the Docker Fields section for more information. |
| **env** | map | Key/value pairs to use as the environment variables for the app and build. |
| **services** | string[] | A list of service instance names to automatically bind to the app. |
| **disk_quota** | quantity | The amount of disk the application should get. Defaults to 1GiB. |
| **memory** | quantity | The amount of RAM to provide the app. Defaults to 1GiB. |
| **cpu** † | quantity | The amount of CPU to provide the application. Defaults to 0.1 (1/10th of a CPU). |
| **instances** | int | The number of instances of the app to run. Defaults to 1. |
| **min-scale** † | int | The minimum number of instances to scale to. Valid only if instances is unset. |
| **max-scale** † | int | The maximum number of instances to scale to. Valid only if instances is unset. Blank means unlimited scaling. |
| **routes** | object | A list of routes the app should listen on. See the Route Fields section for more. |
| **no-route** | boolean | If set to true, the application will not be routable. |
| **random-route** | boolean | If set to true, the app will be given a random route. |
| **timeout** | int | The number of seconds to wait for the app to become healthy. |
| **health-check-type** | string | The type of health-check to use `port`, `none`, or `http`. Default: `port` |
| **health-check-http-endpoint** | string | The endpoint to target as part of the health-check. Only valid if `health-check-type` is `http`. |
| **command** | string | The command that starts the app. If supplied, this will be passed to the container entrypoint. |
| **entrypoint** † | string | Overrides the app container's entrypoint. |
| **args** † | string[] | Overrides the arguments the app container. |

† Unique to Kf


## Docker Fields

The following fields are valid for `application.docker` objects:

| Field | Type | Description |
|:------|:-----|:------------|
| **image** | string | The docker image to use. |

## Route Fields

The following fields are valid for `application.routes` objects:

| Field | Type | Description |
|:------|:-----|:------------|
| **route** | string | A route to the app including hostname, domain, and path. |

## Examples

### Minimal Application

This is a bare-bones manifest that will build an application by auto-detecting
the buildpack based on the uploaded source, and deploy one instance of it.

``` yaml
---
applications:
- name: my-minimal-application
```

### Simple Application

This is a full manifest for a more traditional Java application.

``` yaml
---
applications:
- name: account-manager
  # only upload src/ on push
  path: src
  # use the Java buildpack
  buildpacks:
  - java
  env:
    # manually configure the buildpack's Java version
    BP_JAVA_VERSION: 8
    ENVIRONMENT: PRODUCTION
  # use less disk and memory than default
  disk_quota: 512M
  memory: 512M
  # bump up the CPU
  cpu: 0.2
  instances: 3
  # make the app listen on three routes
  routes:
  - route: accounts.mycompany.com
  - route: accounts.datacenter.mycompany.internal
  - route: mycompany.com/accounts
  # set up a longer timeout and custom endpoint to validate
  # when the app comes up
  timeout: 300
  health-check-type: http
  health-check-http-endpoint: /healthz
  # attach two services by name
  services:
  - customer-database
  - web-cache
```

### Docker Application

Kf can deploy Docker containers as well as manifest deployed applications.
These Docker applications MUST listen on the `PORT` environment variable.

``` yaml
---
applications:
- name: white-label-app
  # use a pre-built docker image (must listen on $PORT)
  docker:
    image: gcr.io/my-company/white-label-app:123
  env:
    # add additional environment variables
    ENVIRONMENT: PRODUCTION
  disk_quota: 1G
  memory: 1G
  cpu: 2
  instances: 1
  routes:
  - route: white-label-app.mycompany.com
```

## Known Differences

The following are known differences between `kf` manifests and `cf` manifests:

* Kf does not support deprecated cf manifest fields. This includes all fields at the root-level of the manifest (other than applications) and routing fields.
* Kf is missing support for the following v2 manifest fields:
  * stack [656](https://github.com/google/kf/issues/656)
  * command [656](https://github.com/google/kf/issues/656)
  * buildpack [656](https://github.com/google/kf/issues/656)
  * docker.username (no support planned)
* Kf is missing support for the following v2 manifest features:
  * Buildpacks from URL [619](https://github.com/google/kf/issues/619)
  * YAML Anchors (no support planned)
  * Manifest variables (no support planned)
* Kf does not yet support v3 manifests. We have planned support for:
  * Metadata [656](https://github.com/google/kf/issues/656)
  * Service parameters [656](https://github.com/google/kf/issues/656)
* Kf does not support auto-detecting ports for Docker containers. (no support planned)
