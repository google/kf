---
title: "kf restart"
weight: 100
description: "Restart each running instance of an App without downtime."
---
### Name

<code translate="no">kf restart</code> - Restart each running instance of an App without downtime.

### Synopsis

<pre translate="no">kf restart APP_NAME [flags]</pre>

### Description

Restarting an App will replace each running instance of an App with a new one.

Instances are replaced one at a time, always ensuring that the desired
number of instances are healthy. This property is upheld by running one
additional instance of the App and swapping it out for an old instance.

The operation completes once all instances have been replaced.


### Examples

<pre translate="no">
kf restart myapp</pre>

### Flags

<dl>
<dt><code translate="no">--async</code></dt>
<dd><p>Do not wait for the action to complete on the server before returning.</p>
</dd>
<dt><code translate="no">-h, --help</code></dt>
<dd><p>help for restart</p>
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


