---
title: "kf scale"
weight: 100
description: "Change the horizontal or vertical scale of an App without downtime."
---
### Name

<code translate="no">kf scale</code> - Change the horizontal or vertical scale of an App without downtime.

### Synopsis

<pre translate="no">kf scale APP_NAME [flags]</pre>

### Description

Scaling an App will change the number of desired instances and/or the
requested resources for each instance.

Instances are replaced one at a time, always ensuring that the desired
number of instances are healthy. This property is upheld by running one
additional instance of the App and swapping it out for an old instance.

The operation completes once all instances have been replaced.


### Examples

<pre translate="no">
# Display current scale settings
kf scale myapp
# Scale to exactly 3 instances
kf scale myapp --instances 3
</pre>

### Flags

<dl>
<dt><code translate="no">--async</code></dt>
<dd><p>Do not wait for the action to complete on the server before returning.</p>
</dd>
<dt><code translate="no">-h, --help</code></dt>
<dd><p>help for scale</p>
</dd>
<dt><code translate="no">-i, --instances=<var translate="no">int32</var></code></dt>
<dd><p>Number of instances, must be &gt;= 1. (default -1)</p>
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


