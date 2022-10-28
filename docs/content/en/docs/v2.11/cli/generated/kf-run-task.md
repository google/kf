---
title: "kf run-task"
weight: 100
description: "Run a short-lived Task on the App."
---
### Name

<code translate="no">kf run-task</code> - Run a short-lived Task on the App.

### Synopsis

<pre translate="no">kf run-task APP_NAME [flags]</pre>

### Description

The run-task sub-command lets operators run a short-lived Task on the App.

### Examples

<pre translate="no">
kf run-task my-app --command &#34;sleep 100&#34; --name my-task</pre>

### Flags

<dl>
<dt><code translate="no">-c, --command=<var translate="no">string</var></code></dt>
<dd><p>Command to execute on the Task.</p>
</dd>
<dt><code translate="no">--cpu-cores=<var translate="no">string</var></code></dt>
<dd><p>Amount of dedicated CPU cores to give the Task (for example 256M, 1024M, 1G).</p>
</dd>
<dt><code translate="no">-k, --disk-quota=<var translate="no">string</var></code></dt>
<dd><p>Amount of dedicated disk space to give the Task (for example 256M, 1024M, 1G).</p>
</dd>
<dt><code translate="no">-h, --help</code></dt>
<dd><p>help for run-task</p>
</dd>
<dt><code translate="no">-m, --memory-limit=<var translate="no">string</var></code></dt>
<dd><p>Amount of dedicated memory to give the Task (for example 256M, 1024M, 1G).</p>
</dd>
<dt><code translate="no">--name=<var translate="no">string</var></code></dt>
<dd><p>Display name to give the Task (auto generated if omitted).</p>
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


