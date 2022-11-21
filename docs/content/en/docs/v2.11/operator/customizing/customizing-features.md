---
title: Customizing Kf Features
description: "Learn to configure your Kf cluster's settings."
weight: 300
---

{{< warning >}}
Changes will be instantly pushed to all Kf resources, test any customization options before using them in production.
{{< /warning >}}

## Build Retention

You can control how many Kf Builds are kept before being garbage collected.

{{< note >}} This example sets the retention to 1 Build. Change the value as needed.{{< /note >}}

<pre class="devsite-terminal devsite-click-to-copy" translate="no">
kubectl patch \
kfsystem kfsystem \
--type='json' \
-p="[{'op': 'replace', 'path': '/spec/kf/config/buildRetentionCount', 'value': <var>1</var>}]"
</pre>

## Task Retention

You can control how many Kf Tasks are kept before being garbage collected.

{{< note >}} This example sets the retention to 1 Task. Change the value as needed.{{< /note >}}

<pre class="devsite-terminal devsite-click-to-copy" translate="no">
kubectl patch \
kfsystem kfsystem \
--type='json' \
-p="[{'op': 'replace', 'path': '/spec/kf/config/taskRetentionCount', 'value': <var>1</var>}]"
</pre>


## Enable or Disable the Istio Sidecar

If you do not require the Istio sidecar for the Build pods, then they can be disabled by setting the value to `true`. Enable by setting the value to `false`.

<pre class="devsite-terminal devsite-click-to-copy" translate="no">
kubectl patch \
kfsystem kfsystem \
--type='json' \
-p="[{'op': 'replace', 'path': '/spec/kf/config/buildDisableIstioSidecar', 'value': <var>true</var>}]"
</pre>

## Enable or Disable Routing Retries

Allows enabling/disbaling retries in the VirtualServices that route traffic to Apps.
Kf leaves this value unset by default and it's inherited from Istio.

Istio's default retry mechanism attempts to make up for instability inherent in service meshes,
however allowing retries requires the contents of the payload to be buffered within Envoy. This
may fail for large payloads and the buffering will need to be disabled at the expense of some
stability.

Values for `routeDisableRetries`:

* `false` Inherit Istio's retry settings. (Default)
* `true` Set retries to 0.

```sh
kubectl patch \
    kfsystem kfsystem \
    --type='json' \
    -p="[{'op':'add','path':'/spec/kf/config/routeDisableRetries','value':true}]"
```

## Build Pod Resource Limits

The default pod resource size can be increased from the default to accommodate very large builds. The units for the value are in `Mi` or `Gi`.

{{< note >}} This is only applicable for built-in Tasks (which is normal for a `kf push` build). For V2 buildpack builds, this will be set on two steps and one for V3 buildpacks or Dockerfiles. This means that for a V2 build the required Pod size will be double the limit. For example, if the memory limit is 1Gi, then the pod will require 2Gi.{{< /note >}}

<pre class="devsite-terminal devsite-click-to-copy" translate="no">
kubectl patch \
kfsystem kfsystem \
--type='json' \
-p="[{'op': 'replace', 'path': '/spec/kf/config/buildPodResources', 'value': {'limits': {'memory': '<var>234Mi</var>'}}}]"
</pre>

Read [Kubernetes container resource docs](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/)  for more information about container resource management.

## Self Signed Certificates for Service Brokers

If you want to use self signed certificates for TLS (`https` instead of `http`) for the service broker URL, the Kf controller requires the CA certificate. To configure Kf for this scenario, create an immutable Kubernetes secret in the `kf` namespace and update the `kfsystem.spec.kf.config.secrets.controllerCACerts.name` object to point to it.

1. Create a secret to store the self-signed certificate.

    {{< note >}}Customize the secret name if desired, or leave the default name of `cacerts`. Replace `/path/to/cert/certs.pem` with the path to the self-signed certificate.{{< /note >}}

    <pre class="devsite-terminal devsite-click-to-copy" translate="no">
    kubectl create secret generic <var>cacerts</var> -nkf --from-file <var>/path/to/cert/certs.pem</var>
    </pre>

1. Make the secret immutable.

    <pre class="devsite-terminal devsite-click-to-copy" translate="no">
    kubectl patch -nkf secret <var>cacerts</var> \
        --type='json' \
        -p="[{'op':'add','path':'/immutable','value':true}]"
    </pre>

1. Update kfsystem to point to the secret.

    {{< note >}}This will cause the controller pod to be re-deployed with the certs mounted as a volume.{{< /note >}}

    ```sh
    kubectl patch \
      kfsystem kfsystem \
      --type='json' \
      -p="[{'op':'add','path':'/spec/kf/config/secrets','value':{'controllerCACerts':{'name':'<var>cacerts</var>'}}}]"
    ```

## Set CPU minimums and ratios

Application default CPU ratios and minimums can be set in the operator.

Values are set in
[CPU units](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#meaning-of-cpu).
Units are typically expressed in millicpus (`m`), or thousandths of a CPU.

The `spec.kf.config.appCPUMin` property specifies a minimum amount of CPU per
application, even if the developer has specified less.

```sh
kubectl patch \
    kfsystem kfsystem \
    --type='json' \
    -p="[{'op':'add','path':'/spec/kf/config/appCPUMin','value':'<var>200m</var>'}]"
```

{{< note >}}Many application runtimes are CPU intensive while initializing
applications. Setting this value too low may cause initial liveness checks to
fail.{{< /note >}}

The `spec.kf.config.appCPUPerGBOfRAM` property specifies a default amount of CPU
to give each app per GB or RAM requested.

You can choose different approaches based on the desired outcome:

*   Choose the ratio of CPU to RAM for the cluster's nodes if you want to
    maximize utilization.
*   Choose a ratio of 1 CPU to 4 GB of RAM which typically works well for I/0 or
    memory bound web applications.

```sh
kubectl patch \
    kfsystem kfsystem \
    --type='json' \
    -p="[{'op':'add','path':'/spec/kf/config/appCPUPerGBOfRAM','value':'<var>250m</var>'}]"
```

## Set buildpacks using git tags

Buildpacks can support pinning by using git tags instead of automatically sourcing the latest buildpack from a git repository.

{{< note >}}Buildpacks can be added using the `kfsystem` operator and `kubectl patch`.{{< /note >}}

Add a new buildpack as follows and use a git tag to specify which version of the buildpack the app should use. Otherwise the buildpack will default to the latest version.

For example, to pin Golang buildpack version 1.9.49 do:

```sh
kubectl patch \
kfsystem kfsystem \
--type='json' \
 -p='[{"op":"add","path":"data/spec/kf/config/spaceBuildpacksV2","value":[{"name":"go_buildpack_v1.9.49","url":"https://github.com/cloudfoundry/go-buildpack.git#v1.9.49"}]}]'
```

This command will add the following to the config-defaults configmaps resource:

```sh
data:
  SpaceBuildpacksV2: |
    - name: go_buildpack_v1.9.49
      url: https://github.com/cloudfoundry/go-buildpack.git#v1.9.49
```

The `kubectl patch` command will replace all the existing buildpacks in the config-defaults configmaps. If the user would like the existing buildpacks to remain, these too need to be included in the command.

To get the list of existing buildpacks in the configmaps run the following command:

```sh
kubectl describe configmaps config-defaults -n kf
```


## Set ProgressDeadlineSeconds

ProgressDeadlineSeconds can be set in the kfsystem operator.


```sh
kubectl patch \
    kfsystem kfsystem \
    --type='json' \
    -p="[{'op':'add','path':'/spec/kf/config/progressDeadlineSeconds','value':100}]"
```
## Set TerminationGracePeriodSeconds

TerminationGracePeriodSeconds can be set in the kfsystem operator.

```sh
kubectl patch \
    kfsystem kfsystem \
    --type='json' \
    -p="[{'op':'add','path':'/spec/kf/config/terminationGracePeriodSeconds','value':200}]"
```

## Set default Kf Task timeout

Kf uses Tekton TaskRuns as its mechanism to run Kf Tasks. 
Tekton may impose a [default timeout on TaskRuns](https://tekton.dev/docs/pipelines/taskruns/#configuring-the-failure-timeout)
that differs depending on the version of Tekton you have installed.

You can override this setting either on the Tekton configmap -- which will set the value for both
Kf Tasks and Builds or on the Kf operator to apply the value only to Tasks.

The following values are supported:

* `null` - Inherit the value from Tekton (Default).
* Value <= 0 - Tasks get an infinite timeout.
* Value >= 0 - Tasks get the given timeout.

Consider setting a long, but non-infinite timeout to prevent improperly programmed
tasks from consuming resources.

```sh
kubectl patch \
    kfsystem kfsystem \
    --type='json' \
    -p="[{'op':'add','path':'/spec/kf/config/taskDefaultTimeoutMinutes','value':-1}]"
```
