---
title: Troubleshoot SourcePackages
---
Use these steps to troubleshoot various issues that can occur when using Kf SourcePackages.

## Object is stuck deleting. {#problem-0}

Run the following command to get the resource information, then check for the causes listed below:

```sh
kubectl get sourcepackages.kf.dev -n SPACE_NAME SOURCEPACKAGE_NAME -o yaml
```

The `kf` CLI can help check for some of the issues:

```sh
kf doctor --space SPACE_NAME sourcepackage/SOURCEPACKAGE_NAME
```

<table>
<thead>
<tr><th>Possible Cause</th><th>Solution</th></tr>
</thead>
<tbody>
<tr>
<td>
Deletion timestamp is in the future.
</td>
<td>
<p>With clock skew the <code>metadata.deletionTimestamp</code> may
still be in the future. Wait a few minutes to see if the object is
deleted.</p>

</td>
</tr>
<tr>
<td>
Finalizers exist on the object.
</td>
<td>
<p>Finalizers are present on the object, they must be
removed by the controller that set them before the object is deleted.</p>

<p>If you want to force a deletion without waiting for the finalizers, edit
the object to remove them from the <code>metadata.finalizers</code> array.</p>

<p>To remove the finalizer from an object, use the
<code>kubectl edit RESOURCE_TYPE RESOURCE_NAME -n my-space</code> command.</p>

<p>See <a href="https://kubernetes.io/blog/2021/05/14/using-finalizers-to-control-deletion/">using finalizers to control deletion</a> to learn more.</p>

<p>Warning: Removing finalizers without allowing the controllers to complete
may cause errors, security issues, data loss, or orphaned resources.</p>

</td>
</tr>
<tr>
<td>
Dependent objects may exist.
</td>
<td>
<p>The object may be waiting on dependents to be deleted before it is deleted.
See the <a href="https://kubernetes.io/docs/concepts/workloads/controllers/garbage-collection/">Kubernetes garbage collection guide to learn more</a>.
Have an administrator check all objects in the namespace and cluster to
see if one of them is blocking deletion.</p>

<p>If you need to remove the object without waiting for dependents, use
<code>kubectl delete</code> with the cascade flag set to: <code>--cascade=orphan</code>.</p>

</td>
</tr>
</tbody>
</table>

## Object generation state drift. {#problem-1}

Run the following command to get the resource information, then check for the causes listed below:

```sh
kubectl get sourcepackages.kf.dev -n SPACE_NAME SOURCEPACKAGE_NAME -o yaml
```

The `kf` CLI can help check for some of the issues:

```sh
kf doctor --space SPACE_NAME sourcepackage/SOURCEPACKAGE_NAME
```

<table>
<thead>
<tr><th>Possible Cause</th><th>Solution</th></tr>
</thead>
<tbody>
<tr>
<td>
Object has generation version drift.
</td>
<td>
<p>This error usually occurs Kf controller did not read the latest version of the object, this
error is usually self-recovered once Kubernetes replicas reach eventual consistency, and it usually does not require
action from users.</p>

</td>
</tr>
</tbody>
</table>

## Object reconciliation failed. {#problem-2}

Run the following command to get the resource information, then check for the causes listed below:

```sh
kubectl get sourcepackages.kf.dev -n SPACE_NAME SOURCEPACKAGE_NAME -o yaml
```

The `kf` CLI can help check for some of the issues:

```sh
kf doctor --space SPACE_NAME sourcepackage/SOURCEPACKAGE_NAME
```

<table>
<thead>
<tr><th>Possible Cause</th><th>Solution</th></tr>
</thead>
<tbody>
<tr>
<td>
Object has TemplateError
</td>
<td>
<p>This error usually occurs if user has entered an invalid property in the custom resource
Spec, or the configuration on the Space/Cluster is bad.</p>

<p>To understand the root cause, user can read the longer error message in the object&rsquo;s <code>status.conditions</code>
using the command:<code>kubectl describe RESOURCE_TYPE RESOURCE_NAME -n space</code>. For example:
<code>kubectl describe serviceinstance my-service -n my-space</code>.</p>

</td>
</tr>
<tr>
<td>
Object has ChildNotOwned error (Name conflicts)
</td>
<td>
<p>This error usually means that the object(s) the controller is trying to create already exists.
This happens if the user created a K8s resource that has the same name as what the controller is trying to create;
but more often it happens if user deletes a resource then Kf controller tries to re-create it. If a child resource
is still hanging around, its owner will be the old resource that no longer exists.</p>

<p>To recover from the error, it is recommended that user deletes the impacted resource and then recreates it. To delete the object,
use a Kf deletion command or use the <code>kubectl delete RESOURCE_TYPE RESOURCE_NAME -n SPACE</code>command. For example,
<code>kf delete-space my-space</code> or <code>kubectl delete space my-space</code>.</p>

<p>To recreate a resource, use a Kf command. For example: <code>kf create-space my-space</code>.</p>

</td>
</tr>
<tr>
<td>
Object has ReconciliationError
</td>
<td>
<p>This error usually means that something has gone wrong with the HTTP call made (by Kf controller)
to the Kubernetes API servier to create/update resource.</p>

<p>To understand the root cause, user can read the longer error message in the object&rsquo;s <code>status.conditions</code>
using the command:<code>kubectl describe RESOURCE_TYPE RESOURCE_NAME -n space</code>. For example:
<code>kubectl describe serviceinstance my-service -n my-space</code>.</p>

</td>
</tr>
</tbody>
</table>


