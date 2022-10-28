---
title: "kf delete-space"
weight: 100
description: "Delete the Space with the given name."
---
### Name

<code translate="no">kf delete-space</code> - Delete the Space with the given name.

### Synopsis

<pre translate="no">kf delete-space NAME [flags]</pre>

### Description

Deletes the Space with the given name and wait for it to be deleted.

Kubernetes will delete the Space once all child resources it owns have been deleted.
Deletion may take a long time if any of the following conditions are true:

* There are many child objects.
* There are finalizers on the object preventing deletion.
* The cluster is in an unhealthy state.

Deleting a Space will also delete its:

* Apps
* Build history
* Service bindings
* Service instances
* Routes
* The backing Kubernetes Namespace
* Kubernetes resources in the Namespace

You will be unable to make changes to resources in the Space once deletion
has begun.

### Examples

<pre translate="no">
kf delete-space my-space</pre>

### Flags

<dl>
<dt><code translate="no">--async</code></dt>
<dd><p>Do not wait for the action to complete on the server before returning.</p>
</dd>
<dt><code translate="no">-h, --help</code></dt>
<dd><p>help for delete-space</p>
</dd>
<dt><code translate="no">--retries=<var translate="no">int</var></code></dt>
<dd><p>Number of times to retry execution if the command isn't successful. (default 5)</p>
</dd>
<dt><code translate="no">--retry-delay=<var translate="no">duration</var></code></dt>
<dd><p>Set the delay between retries. (default 1s)</p>
</dd>
</dl>


### Inherited flags

These flags are inherited from parent commands.

<dl>
<dt><code translate="no">--as=<var translate="no">string</var></code></dt>
<dd><p>Username to impersonate for the operation.</p>
</dd>
<dt><code translate="no">--as-group=<var translate="no">strings</var></code></dt>
<dd><p>Group to impersonate for the operation. Include this flag multiple times to specify multiple groups.</p>
</dd>
<dt><code translate="no">--config=<var translate="no">string</var></code></dt>
<dd><p>Path to the Kf config file to use for CLI requests.</p>
</dd>
<dt><code translate="no">--kubeconfig=<var translate="no">string</var></code></dt>
<dd><p>Path to the kubeconfig file to use for CLI requests.</p>
</dd>
<dt><code translate="no">--log-http</code></dt>
<dd><p>Log HTTP requests to standard error.</p>
</dd>
<dt><code translate="no">--space=<var translate="no">string</var></code></dt>
<dd><p>Space to run the command against. This flag overrides the currently targeted Space.</p>
</dd>
</dl>


