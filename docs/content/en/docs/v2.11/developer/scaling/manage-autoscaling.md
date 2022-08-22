---
title: Manage Autoscaling
description: "Learn to use autoscaling for your app."
weight: 10
---

Kf Apps can be automatically scaled based on CPU usage.
You can configure autoscaling limits for your Apps and the target CPU usage for
each App instance. Kf automatically scales your Apps up
and down in response to demand.

By default, autoscaling is disabled. Follow the steps below to enable autoscaling.

## View Apps

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

## Update autoscaling limits

You can update the instance limits using the `kf update-autoscaling-limits`
command.

```sh
kf update-autoscaling-limits app-name min-instances max-instances
```

## Create autoscaling rule

You can create autoscaling rules using the `kf create-autoscaling-rule`
command.

```sh
kf create-autoscaling-rule app-name CPU min-threshold max-threshold
```

## Delete autoscaling rules

You can delete all autoscaling rules with the
`kf delete-autoscaling-rule` command. Kf only supports
one autoscaling rule.

```sh
kf delete-autoscaling-rules app-name
```

## Enable and disable autoscaling

Autoscaling can be enabled by using `enable-autoscaling` and
disabled by using `disable-autoscaling`. When it is disabled, the
configurations, including limits and rules, are preserved.

```sh
kf enable-autoscaling app-name

kf disable-autoscaling app-name
```