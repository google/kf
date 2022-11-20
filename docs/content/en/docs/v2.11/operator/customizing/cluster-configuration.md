---
title: Cluster Configuration 
description: "Learn to configure your Kf cluster's settings."
weight: 200
---

{{< warning >}}
Changes will be instantly pushed to all Kf resources, test any customization options before using them in production.
{{< /warning >}}

Kf uses a Kubernetes configmap named `config-defaults` in
the `kf` namespace to store cluster wide configuration settings.
This document explains its structure and fields.

{{< note >}}

This configuration is usually copied from the `kfsystem` by the Kf operator.
The values can be set in the operator under the `spec.kf.config` path.

{{< /note >}}

## Structure of the config-defaults configmap

The configmap contains three types of key/value pairs in the `.data` field:

*  Comment keys prefixed by `_` contain examples, notes, and warnings.
*  String keys contain plain text values.
*  Object keys contain a JSON or YAML value that has been encoded as a string.

Example:

```yaml
_note: "This is some note"
stringKey: "This is a string key that's not encoded as JSON or YAML."
objectKey: |
  - "These keys contain nested YAML or JSON."
  - true
  - 123.45
```

{{< note >}} Kubernetes requires all configmap values to be strings.
Kf validates the structure of the fields before the
object is updated.{{< /note >}}

## Example section

The example section under the `_example` key contains explanations for other
fields and examples. Changes to this section have no effect.

## Space container registry

The `spaceContainerRegistry` property is a plain text value that specifies the
default container registry each space uses to store built images.

Example:

```yaml
spaceContainerRegistry: gcr.io/my-project
```

## Space cluster domains

The `spaceClusterDomains` property is a string encoded YAML array of domain objects.

Each space in the cluster adds all items in the array to its list of
domains that developers can bind their apps to.

<table class="properties responsive">
  <thead>
    <tr>
      <th colspan="2">Fields</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td><code>domain</code></td>
      <td>
        <p><code class="apitype">string</code></p>
        <p>The domain name to make available. May contain one of the following substitutions:</p>
        <ul>
          <li><code>$(SPACE_NAME)</code> - Replaced in each space with the name of the space.</li>
          <li><code>$(CLUSTER_INGRESS_IP)</code> - The IP address of the cluster ingress gateway.</li>
        </ul>
      </td>
    </tr>
    <tr>
      <td><code>gatewayName</code></td>
      <td>
        <p><code class="apitype">string</code></p>
        <p>(Optional)</p>
        <p>
          Overrides the Istio gateway routes will be bound to.
          Defaults to <code>kf/external-gateway</code>, but any
          other gateway in the <code>kf</code> namespace may be used.
        </p>
      </td>
    </tr>
  </tbody>
</table>


{{< note >}} All domains should contain the `$(SPACE_NAME)` substitution to prevent
Apps in different namespaces from listening to the same hostname which causes
undefined behavior.{{< /note >}}

Example:

```yaml
spaceClusterDomains: |
  # Support canonical and vanity domains
  - domain: $(SPACE_NAME).prod.example.com
  - domain: $(SPACE_NAME).kf.us-east1.prod.example.com

  # Using a dynamic DNS resolver
  - domain: $(SPACE_NAME).$(CLUSTER_INGRESS_IP).nip.io

  # Creating an internal domain only visible within the cluster
  - domain: $(SPACE_NAME)-apps.internal
    gatewayName: kf/internal-gateway
```

## Buildpacks V2 lifecycle builder

The `buildpacksV2LifecycleBuilder` property contains the version of the Cloud Foundry
`builder` binary used execute buildpack v2 builds.

The value is a Git reference. To use a specific version, append an `@` symbol
followed by a Git SHA to the end.

Example:

```yaml
buildpacksV2LifecycleBuilder: "code.cloudfoundry.org/buildpackapplifecycle/builder@GIT_SHA"
```

## Buildpacks V2 lifecycle launcher

The `buildpacksV2LifecycleLauncher` property contains the version of the Cloud Foundry
`launcher` binary built into every buildpack V2 application.

The value is a Git reference. To use a specific version, append an `@` symbol
followed by a Git SHA to the end.

Example:

```yaml
buildpacksV2LifecycleLauncher: "code.cloudfoundry.org/buildpackapplifecycle/launcher@GIT_SHA"
```

## Buildpacks V2 list

The `spaceBuildpacksV2` property is a string encoded YAML array that holds an ordered
list of default buildpacks that are used to build applications compatible with
the V2 buildpacks process.

<table class="properties responsive">
  <thead>
    <tr>
      <th colspan="2">Fields</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td><code>name</code></td>
      <td>
        <p><code class="apitype">string</code></p>
        <p>A short name developers can use to reference the buildpack by in their application manifests.</p>
      </td>
    </tr>
    <tr>
      <td><code>url</code></td>
      <td>
        <p><code class="apitype">string</code></p>
        <p>The URL used to fetch the buildpack.</p>
      </td>
    </tr>
    <tr>
      <td><code>disabled</code></td>
      <td>
        <p><code class="apitype">boolean</code></p>
        <p>Used to prevent this buildpack from executing.</p>
      </td>
    </tr>
  </tbody>
</table>


## Stacks V2 list

The `spaceBuildpacksV2` property is a string encoded YAML array that holds an
ordered list of stacks that can be used with Cloud Foundry compatible builds.

<table class="properties responsive">
  <thead>
    <tr>
      <th colspan="2">Fields</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td><code>name</code></td>
      <td>
        <p><code class="apitype">string</code></p>
        <p>A short name developers can use to reference the stack by in their application manifests.</p>
      </td>
    </tr>
    <tr>
      <td><code>image</code></td>
      <td>
        <p><code class="apitype">string</code></p>
        <p>
          URL of the container image to use as the stack.
          For more information, see <a href="https://kubernetes.io/docs/concepts/containers/images">https://kubernetes.io/docs/concepts/containers/images</a>.
        </p>
      </td>
    </tr>
  </tbody>
</table>

{{< note >}} Use images tagged with SHAs to improve caching.{{< /note >}}

## Stacks V3 list

The `spaceStacksV3` property is a string encoded YAML array that holds an ordered
list of stacks that can be used with
[Cloud Native Buildpack](https://buildpacks.io/)
builds.

<table class="properties responsive">
  <thead>
    <tr>
      <th colspan="2">Fields</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td><code>name</code></td>
      <td>
        <p><code class="apitype">string</code></p>
        <p>A short name developers can use to reference the stack by in their application manifests.</p>
      </td>
    </tr>
    <tr>
      <td><code>description</code></td>
      <td>
        <p><code class="apitype">string</code></p>
        <p>A short description of the stack shown when running <code>kf stacks</code>.</p>
      </td>
    </tr>
    <tr>
      <td><code>buildImage</code></td>
      <td>
        <p><code class="apitype">string</code></p>
        <p>
          URL of the container image to use as the builder.
          For more information, see <a href="https://kubernetes.io/docs/concepts/containers/images">https://kubernetes.io/docs/concepts/containers/images</a>.
        </p>
      </td>
    </tr>
    <tr>
      <td><code>runImage</code></td>
      <td>
        <p><code class="apitype">string</code></p>
        <p>
          URL of the container image to use as the base for all apps built with .
          For more information, see <a href="https://kubernetes.io/docs/concepts/containers/images">https://kubernetes.io/docs/concepts/containers/images</a>.
        </p>
      </td>
    </tr>
    <tr>
      <td><code>nodeSelector</code></td>
      <td>
        <p><code class="apitype">map (key: string, value: string)</code></p>
        <p>(Optional)</p>
        <p>
          A NodeSelector used to indicate which nodes applications built with
          this stack can run on.
        </p>
      </td>
    </tr>
  </tbody>
</table>

{{< note >}} Use images tagged with SHAs to improve caching.{{< /note >}}

Example:

```yaml
spaceStacksV3: |
  - name: heroku-18
    description: The official Heroku stack based on Ubuntu 18.04
    buildImage: heroku/pack:18-build
    runImage: heroku/pack:18
    nodeSelector:
       kubernetes.io/os: windows
```

## Default to V3 Stack

The `spaceDefaultToV3Stack` property contains a quoted value `true` or `false`
indicating whether spaces should use V3 stacks if a user doesn't specify one.

## Feature flags

The `featureFlags` property contains a string encoded YAML map of feature flags
that can enable and disable features of Kf.

Flag names that aren't supported by Kf will be ignored.

| Flag Name                   | Default | Purpose |
| ---                         | ---     | --- |
| `disable_custom_builds`     | `false` | Disable developer access to arbitrary Tekton build pipelines. |
| `enable_dockerfile_builds`  | `true`  | Allow developers to build source code from dockerfiles. |
| `enable_custom_buildpacks`  | `true`  | Allow developers to specify external buildpacks in their applications. |
| `enable_custom_stacks`      | `true`  | Allow developers to specify custom stacks in their applications. |

Example:

```yaml
featureFlags: |
  disable_custom_builds: false
  enable_dockerfile_builds: true
  enable_some_feature: true
```
## ProgressDeadlineSeconds

`ProgressDeadlineSeconds` contains a configurable quoted integer indicating the maximum allowed time between state transition and reaching a stable state before provisioning or deprovisioning when pushing an application. The default value is `600` seconds.

## TerminationGracePeriodSeconds

The `TerminationGracePeriodSeconds` contains a configurable quoted integer indicating the time between when the processes running in the pod are sent a termination signal and the time when the processes are forcibly halted with a kill signal. The default value is `30` seconds.

