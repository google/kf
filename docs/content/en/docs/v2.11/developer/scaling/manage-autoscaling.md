---
title: Manage Autoscaling
description: "Learn to use autoscaling for your app."
weight: 10
---

Kf supports two primary autoscaling modes:

* [Built-in autosacling similar to Cloud Foundry](#built-in-autoscaling).
* [Advanced autoscaling through the Kubernetes Horizontal Pod Autoscaler (HPA)](#advanced-autoscaling).

## Built-in autoscaling {#built-in-autoscaling}

Kf Apps can be automatically scaled based on CPU usage.
You can configure autoscaling limits for your Apps and the target CPU usage for
each App instance. Kf automatically scales your Apps up
and down in response to demand.

By default, autoscaling is disabled. Follow the steps below to enable autoscaling.

### View Apps

You can view the autoscaling status for an App using the `kf apps`
command. If autoscaling is enabled for an App, `Instances` includes the
autoscaling status.

```none
$ kf apps

Name   Instances              Memory  Disk  CPU
app1   4 (autoscaled 4 to 5)  256Mi   1Gi   100m
app2   1                      256Mi   1Gi   100m
```
Autoscaling is enabled for `app1` with `min-instances` set to 4 and
`max-instances` set to 5. Autoscaling is disabled for `app2`.

### Update autoscaling limits

You can update the instance limits using the `kf update-autoscaling-limits`
command.

```sh
kf update-autoscaling-limits app-name min-instances max-instances
```

### Create autoscaling rule

You can create autoscaling rules using the `kf create-autoscaling-rule`
command.

```sh
kf create-autoscaling-rule app-name CPU min-threshold max-threshold
```

### Delete autoscaling rules

You can delete all autoscaling rules with the
`kf delete-autoscaling-rule` command. Kf only supports
one autoscaling rule.

```sh
kf delete-autoscaling-rules app-name
```

### Enable and disable autoscaling

Autoscaling can be enabled by using `enable-autoscaling` and
disabled by using `disable-autoscaling`. When it is disabled, the
configurations, including limits and rules, are preserved.

```sh
kf enable-autoscaling app-name

kf disable-autoscaling app-name
```

## Advanced autoscaling {#advanced-autoscaling}

Kf Apps support the Kubernetes Horizontal Pod Autoscaler interface and will
therefore work with HPAs created using `kubectl`.

Kubernetes HPA policies are less restrictive than Kf's built-in support for autoscaling.

They include support for:

* Scaling on memory, CPU, or disk usage.
* Scaling based on custom metrics, such as traffic load or queue length.
* Scaling on multiple metrics.
* The ability to tune reactivity to smooth out rapid scaling.

### Using custom HPAs with apps

You can follow the [Kubernetes HPA walkthrough](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale-walkthrough/) to learn how to set up autoscalers.

When you create the HPA, make sure to set the `scaleTargetRef` to be your application:

{{< highlight yaml "hl_lines=7-10">}}
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: app-scaler
  namespace: SPACE_NAME
spec:
  scaleTargetRef:
    apiVersion: kf.dev/v1alpha1
    kind: App
    name: APP_NAME
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 60
{{</ highlight >}}

### Caveats

* You shouldn't use Kf autoscaling with an HPA.
* When you use an HPA, `kf apps` will show the current number of instances, it won't show that the App
  is being autoscaled.

