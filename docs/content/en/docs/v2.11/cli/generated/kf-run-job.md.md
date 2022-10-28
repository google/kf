---
title: "kf run-job"
weight: 100
description: "Run the Job once."
---
### Name

<code translate="no">kf run-job</code> - Run the Job once.

### Synopsis

<pre translate="no">kf run-job JOB_NAME [flags]</pre>

### Description

The run-job sub-command lets operators run a Job once.

### Examples

<pre translate="no">
kf run-job my-job</pre>

### Flags

<dl>
<dt><code translate="no">--cpu-cores=<var translate="no">string</var></code></dt>
<dd><p>Amount of dedicated CPU cores to give the Task (for example 256M, 1024M, 1G).</p>
</dd>
<dt><code translate="no">-k, --disk-quota=<var translate="no">string</var></code></dt>
<dd><p>Amount of dedicated disk space to give the Task (for example 256M, 1024M, 1G).</p>
</dd>
<dt><code translate="no">-h, --help</code></dt>
<dd><p>help for run-job</p>
</dd>
<dt><code translate="no">-m, --memory-limit=<var translate="no">string</var></code></dt>
<dd><p>Amount of dedicated memory to give the Task (for example 256M, 1024M, 1G).</p>
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


