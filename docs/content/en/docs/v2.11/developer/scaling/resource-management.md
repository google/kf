---
title: Managing Resources for Apps
description: "Learn to set resources on apps."
weight: 20
---

When you create an app, you can optionallyl specify how much of each resource each instance
of the application will receive when it runs.

Kf simplifies the Kubernetes model of resources and provides defaults that should work for
most I/O bound appliations out of the box.

## Resource types

Kf supports three types of resources, memory, CPU, and ephemeral disk.

* Memory specifies the amount of RAM an application receives when running. If it exceeds this amount
  then the container is restarted.
* Ephemeral disk specifies how much an application can write to a local disk. If an application exceeds
  this amount then it may not be able to write more.
* CPU specifies the number of CPUs an application receives when running.

### Manifest

Resources are specified using four values in the manifest:

* `memory` sets the guaranteed minimum an app will receive and the maximum it's permitted to use.
* `disk_quota` sets the guaranteed minimum an app will receive and the maximum it's permitted to use.
* `cpu` sets the guaranteed minimum an app will receive.
* `cpu-limit` sets the maximum CPU an app can use.

Example:

```yaml
applications:
- name: "example"
  disk_quota: 512M
  memory: 512M
  cpu: 200m
  cpu-limit: 2000m
```

### Defaults

Memory and ephemeral storage are both set to 1Gi if not specified.

CPU defautls to one of the following

* 1/10th of a CPU if the platform operator hasn't overridden it.
* A CPU value proportionally scaled by the amount of memory requested.
* A minimum CPU value set by the platform operator.

## Resource units

### Memory and disk

Cloud Foundry used the units `T`, `G`, `M`, and `K` to represent powers of two. 
Kubernetes uses the units `Ei`, `Pi`, `Gi`, `Mi`, and `Ki` for the same.

Kf allows you to specify memory and disk in either units.

### CPU

Kf and Kubernetes use the unit `m` for CPU, representing milli-CPU cores (thousandths of a core).

## Sidecar overhead

When Kf schedules your app's container as a Kubernetes Pod, it may bundle additional containers to your app
to provide additional functionality. It's likely your application will also have an Istio sidecar which is
responsible for networking.

These containers will supply their own resource requests and limits and are overhead associated with running
your application.

## Best practices

* All applications should set memory and disk quotas.
* CPU intensive applications should set a CPU and limit to guarantee they'll have the resources they need without
  starving other apps.
* I/O bound applications shouldn't set a CPU limit so they can burst during startup.

## Additional reading

* Learn how [Kubernetes defines resources and schedules apps based on resource use](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/)