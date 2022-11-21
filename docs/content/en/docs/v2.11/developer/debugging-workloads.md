---
title:  Debugging workloads
description: "Understand how to debug workloads using Kubernetes."
weight: 200
---

Kf uses Kubernetes under the hood to schedule application workloads onto a cluster.
Ultimately, every workload running on a Kubernetes cluster is scheduled as a Pod,
but the Pods may have different properties based on the higher level abstractions
that schedule them.

Once you understand how to debug Pods, you can debug any running workload on the cluster
including the components that make up Kf and Kubernetes.

## General debugging

Kubernetes divides most resources into Namespaces. Each Kf Space creates a Namespace
with the same name. If you're debugging a Kf resource, you'll want to remember to set
the `-n` or `--namespace` CLI flag on each `kubectl` command you run.

{{< note >}}
This guide is a good jumping off point, but you should also look at
Kubernetes' [extensive documentation](https://kubernetes.io/docs/tasks/debug/debug-application/)
about troubleshooting.
{{< /note >}}

### Finding a Pod

You can list all the Pods in a Namespace using the command:

```sh
kubectl get pods -n NAMESPACE
```

This will list all Pods in the Namespace. You'll often want to filter these
using a [label selector](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/)
like the following:

```sh
kubectl get pods -n NAMESPACE -l "app.kubernetes.io/component=app-server,app.kubernetes.io/name=echo"
```

You can get the definition of a Pod using a command like the following:

```sh
kubectl get pods -n NAMESPACE POD_NAME -o yaml
```

{{< note >}}
`kubectl` supports a variety of output types similar to Kf using the `-o` flag.
`yaml`, `json`, and `jsonpath` are common ones.
{{< /note >}}

### Understand Kubernetes objects

Most Kubernetes objects follow the same general structure. Kubernetes also has extensive
help documentation for each type of object (including Kf's).

If you need to quickly look up what a field is for, use the `kubectl explain` command
with the object type and the path to the field you want documentation on.

```
$ kubectl explain pod.metadata.labels
KIND:     Pod
VERSION:  v1

FIELD:    labels <map[string]string>

DESCRIPTION:
     Map of string keys and values that can be used to organize and categorize
     (scope and select) objects. May match selectors of replication controllers
     and services. More info: http://kubernetes.io/docs/user-guide/labels
```

When you run `kubectl get RESOURCE_TYPE RESOURCE_NAME -oyaml` you'll see the stored version
of the object. An annotated example of an Pod running the instance of an App is below:

```yaml
apiVersion: v1
kind: Pod
metadata:
    # Annotations hold security information and configuration for the 
    # object.
    annotations:
        kubectl.kubernetes.io/default-container: user-container
        kubectl.kubernetes.io/default-logs-container: user-container
        sidecar.istio.io/inject: "true"
        traffic.sidecar.istio.io/includeOutboundIPRanges: '*'
    # Labels hold information useful for filtering resources.
    labels:
        # Kf sets many of these labels to help find Pods.
        app.kubernetes.io/component: app-server
        app.kubernetes.io/managed-by: kf
        app.kubernetes.io/name: echo
        kf.dev/networkpolicy: app
    name: echo-6b759c978b-zwrt8
    namespace: development
    # Contains the object(s) that "own" this resource, this usually
    # means the ones that were responsible for creating it.
    ownerReferences:
    - apiVersion: apps/v1
        blockOwnerDeletion: true
        controller: true
        kind: ReplicaSet
        name: echo-6b759c978b
        uid: 7f0ee42d-f1e8-4c4f-b0c8-f0c5d7f27a0a
    # The ID of the resource, if deleted and re-created the ID will
    # change.
    uid: 0d49d5d7-afa4-4904-9f69-f98ce1923745
spec:
    # Contains the desired state of the object, written by a human or
    # one of the metadata.ownerReferences.
    # Omitted for brevity.
status:
    # Contains the state of the object as written by the controller.
    # Omitted for brevity.
```

When you see an object with an `metadata.ownerReferences` set, you can run
`kubectl get` again to find that object's information all the way back up
until you find the root object responsible for creating it. In this case
the chain would look like Pod -> ReplicaSet -> Deployment -> App.


### Getting logs

You can get logs for a specific Pod using `kubectl logs`:

```
kubectl logs -c user-container -n NAMESPACE POD_NAME
```

You can find a list of containers on the Pod in the `.spec.containers.name` field:

```
kubectl get pods -n NAMESPACE -o jsonpath='{.spec.containers[*].name}' POD_NAME
```

{{< note >}}
Kf calls the container running the App code `user-container`.
On newer versions of Kf, you won't have to supply `-c` due to
the `kubectl.kubernetes.io/default-container` annotation on the
Pod that `kubectl` knows how to read.
{{< /note >}}

### Port forwarding

You can [port-forward](https://kubernetes.io/docs/tasks/access-application-cluster/port-forward-access-application-cluster/)
to a specific port using `kubectl port-forward`.

```sh
# Bind remote port 8080 to local port 8080.
kubectl port-forward -n NAMESPACE POD_NAME 8080

# Bind remote port 8080 to local port 9000.
kubectl port-forward -n NAMESPACE POD_NAME 9000:8080
```

### Open a shell

You can [open a shell to a container](https://kubernetes.io/docs/tasks/debug/debug-application/get-shell-running-container/)
using `kubectl exec`.

```sh
kubectl exec --stdin --tty -n NAMESPACE -c user-container POD_NAME -- /bin/bash
```

{{< note >}}
Kf calls the container running the App code `user-container`.
On newer versions of Kf, you won't have to supply `-c` due to
the `kubectl.kubernetes.io/default-container` annotation on the
Pod that `kubectl` knows how to read.
{{< /note >}}

## Kf Specifics

The following sections show specific information about Kf types that schedule Pods.

### App Pods

**Label selector**

All Apps:

`app.kubernetes.io/component=app-server,app.kubernetes.io/managed-by=kf`

Specific App (replace `APP_NAME`):

`app.kubernetes.io/component=app-server,app.kubernetes.io/managed-by=kf,app.kubernetes.io/name=APP_NAME`


**Expected containers**

* `user-container` Container running the application code.
* `istio-proxy` Connects the application code to the virtual network.

**Ownership hierarchy**

Each resource will have a `metadata.ownerReferences` to the resource below it:

1. Kubernetes **Pod** Runs a single instance of the application code.
1. Kubernetes **ReplicaSet** Schedules all the instances for one version of an App. 
1. Kubernetes **Deployment** Manages rollouts and scaling of multiple versions of the App.
1. Kf **App** Orchestrates routing, rollouts, service bindings, builds, etc. for an App.

### Build Pods

**Label selector**

All Builds:

`app.kubernetes.io/component=build,app.kubernetes.io/managed-by=kf`

Builds for a specific App (replace `APP_NAME`):

`app.kubernetes.io/component=build,app.kubernetes.io/managed-by=kf,app.kubernetes.io/name=APP_NAME`

Specific Build for an App (replace `APP_NAME` and `TASKRUN_NAME`):

`app.kubernetes.io/component=build,app.kubernetes.io/managed-by=kf,app.kubernetes.io/name=APP_NAME,tekton.dev/taskRun=TASKRUN_NAME`

**Expected containers**

* `step-.*` Container(s) that execute different steps of the build process.
  Specific steps depend on the type of build.
* `istio-proxy` (Optional) Connects the application code to the virtual network.

**Ownership hierarchy**

Each resource will have a `metadata.ownerReferences` to the resource below it:

1. Kubernetes **Pod** Runs the steps of the build.
1. Tekton **TaskRun** Schedules the Pod, ensures it runs to completion once, and cleans it up once done.
1. Kf **Build** Creates a TaskRun with the proper steps to build an App from source.
1. Kf **App** Creates Builds with app-specific information like environment variables.


### Task Pods

**Label selector**

All Tasks:

`app.kubernetes.io/component=task,app.kubernetes.io/managed-by=kf`

Tasks for a specific App (replace `APP_NAME`):

`app.kubernetes.io/component=task,app.kubernetes.io/managed-by=kf,app.kubernetes.io/name=APP_NAME`

Specific Task for an App (replace `APP_NAME` and `TASKRUN_NAME`):

`app.kubernetes.io/component=task,app.kubernetes.io/managed-by=kf,app.kubernetes.io/name=APP_NAME,tekton.dev/taskRun=TASKRUN_NAME`

**Expected containers**

* `step-user-container` Container running the application code.
* `istio-proxy` Connects the application code to the virtual network.

**Ownership hierarchy**

1. Kubernetes **Pod** Runs an instance of the application code.
1. Tekton **TaskRun** Schedules the Pod, ensures it runs to completion once, and cleans it up once done.
1. Kf **Task** Creates a TaskRun with the proper step to run a one-off Task.
1. Kf **App** Creates Builds with app-specific information like environment variables.

## Next steps

* Explore the Kubernetes [guide to debugging applications](https://kubernetes.io/docs/tasks/debug/debug-application/).
* Understand how to write [label selectors](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/).