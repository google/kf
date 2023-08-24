---
title: "Scaling overview"
linkTitle: "Overview"
description: "Learn about how Kf scales apps."
weight: 1
---

Kf leverages the Kubernetes Horizontal Pod Autoscaler (HPA)
to automatically scale the number of Pods in a App. When autoscaling is enabled
for an App, an HPA object is created and bound to the App object. It then dynamically
calculates the target scale and sets it for the App.

Kf Apps are also compatible with HPA policies created outside of Kf.

## How Kf scaling works

The number of Pods that are deployed for a Kf App is
controlled by its underlying Deployment object's `replicas` field. The target
number of Deployment replicas is set through the App's `replicas` field.

Scaling can be done manually with the `kf scale` command. 
This command is disabled when autoscaling is enabled to avoid conflicting targets.

## How the Kubernetes Horizontal Pod Autoscaler works

The Horizontal Pod Autoscaler (HPA) is implemented as a Kubernetes API resource
(the HPA object) and a control loop (the HPA controller) which periodically
calculates the number of desired replicas based on current resource utilization.
The HPA controller then passes the number to the target object that implements the
Scale subresource. The actual scaling is delegated to the underlying object and
its controller. You can find more information in the [Kubernetes documentation](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/).

### How the Autoscaler determines when to scale

Periodically, the HPA controller queries the resource utilization against
the metrics specified in each HorizontalPodAutoscaler definition. The controller
obtains the metrics from the resource metrics API for each Pod. Then the
controller calculates the utilization value as a percentage of the equivalent
resource request. The desired number of replicas is then calculated based on the
ratio of current percentage and desired percentage. You can read more about the
[autoscaling algorithm](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/#algorithm-details) in the Kubernetes documentation.

### Metrics

Kf uses HPA v1 which only supports CPU as the target metric.


## How the Kubernetes Horizontal Autoscaler works with Kf

When autoscaling is enabled for a Kf App, the Kf
controller will create an HPA object based on the scaling limits and rules
specified on the App. Then the HPA controller fetches the specs from the HPA
object and scales the App accordingly.

The HPA object will be deleted if Autoscaling is disabled or if the
corresponding App is deleted.
