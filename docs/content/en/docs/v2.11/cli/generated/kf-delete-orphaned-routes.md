---
title: "kf delete-orphaned-routes"
weight: 100
description: "Delete Routes with no App bindings."
---
### Name

<code translate="no">kf delete-orphaned-routes</code> - Delete Routes with no App bindings.

### Synopsis

<pre translate="no">kf delete-orphaned-routes [flags]</pre>

### Description

Deletes Routes in the targeted Space that are fully reconciled and
don't have any bindings to Apps.

### Examples

<pre translate="no">
kf delete-orphaned-routes</pre>

### Flags

<dl>
<dt><code translate="no">-h, --help</code></dt>
<dd><p>help for delete-orphaned-routes</p>
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


