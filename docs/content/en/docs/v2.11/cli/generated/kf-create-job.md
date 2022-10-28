---
title: "kf create-job"
weight: 100
description: "Create a Job on the App."
---
### Name

<code translate="no">kf create-job</code> - Create a Job on the App.

### Synopsis

<pre translate="no">kf create-job APP_NAME JOB_NAME COMMAND [flags]</pre>

### Description

The create-job sub-command lets operators create a Job that can be run on a schedule or ad hoc.

### Examples

<pre translate="no">
kf create-job my-app my-job &#34;sleep 100&#34;</pre>

### Flags

<dl>
<dt><code translate="no">--async</code></dt>
<dd><p>Do not wait for the action to complete on the server before returning.</p>
</dd>
<dt><code translate="no">-c, --concurrency-policy=<var translate="no">string</var></code></dt>
<dd><p>Specifies how to treat concurrent executions of a Job: Always (default), Replace, or Forbid. (default &quot;Always&quot;)</p>
</dd>
<dt><code translate="no">--cpu-cores=<var translate="no">string</var></code></dt>
<dd><p>Amount of dedicated CPU cores to give the Task (for example 256M, 1024M, 1G).</p>
</dd>
<dt><code translate="no">-k, --disk-quota=<var translate="no">string</var></code></dt>
<dd><p>Amount of dedicated disk space to give the Task (for example 256M, 1024M, 1G).</p>
</dd>
<dt><code translate="no">-h, --help</code></dt>
<dd><p>help for create-job</p>
</dd>
<dt><code translate="no">-m, --memory-limit=<var translate="no">string</var></code></dt>
<dd><p>Amount of dedicated memory to give the Task (for example 256M, 1024M, 1G).</p>
</dd>
<dt><code translate="no">-s, --schedule=<var translate="no">string</var></code></dt>
<dd><p>Cron schedule on which to execute the Job.</p>
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


