---
title: "Temporarily stop reconciliation"
description: "Stop resource reconciliation to change values for debugging."
---

Sometimes it may be necessary to stop the Kf controller or Kf operator from
changing settings on a Kubernetes resource if you need to change values on it
for debugging.


## Operator managed resources

Resources managed by the Kf operator, usually found in the `kf` namespace,
can have reconciliation disabled using the `operator.knative.dev/mode` annotation:


| Value                | Behavior                                                                           |
| -------------------- | ---------------------------------------------------------------------------------- |
| `EnsureExists`       | The resource will be created if it doesn't exist, but values won't be overwritten. |
| `Reconcile` or blank | The resource will be created if it doesn't exist, and values will be overwritten.  |

To change values on operator managed resources for testing:

1. Make a local copy of the resource so you can revert it later.
2. Disable reconciliation by setting the annotation to `EnsureExists`.

    Example disabling reconciliation on the Kf controller

    ```shell
    kubectl annotate --overwrite -n kf deployment controller operator.knative.dev/mode=EnsureExists
    ```
3. Update the resource as needed for testing.
4. When done, restore the annotation to the original value and optionally restore the changed values.


## Kf managed resources

Kf managed resources are child resources created in response to configuration set on Spaces, Apps,
Builds, etc. These usually have the label `app.kubernetes.io/managed-by: kf`.

You can disable reconciliation by removing the `metadata.ownerReference` that references the
parent Kf resource. Save a copy of the field so you can add it back when you're done.

{{< warning >}}
The owning Kf resource will report an error when the ownership is removed.
Kf may also stop reconciling other child resources if they depend on the resource
you've modified.
{{< /warning >}}

1. Make a local copy of the resource so you can revert it later.
2. Disable reconciliation by removing the `metadata.ownerReference` field.

    Example editing a Namespaced owned by a Kf Space:

    ```shell
    $ kubectl edit namespace test
    ```

    Empty the `metadata.ownerReference` field:

    ```yaml
    apiVersion: v1
    kind: Namespace
    metadata:
        creationTimestamp: "2023-04-19T20:34:50Z"
        name: test
        ownerReferences: []
    ```

3. Update the resource as needed for testing.
4. When done, restore the resource to the original value.
