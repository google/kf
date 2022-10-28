---
title: "kf schedule-job"
weight: 100
description: "Schedule the Job for execution on a cron schedule."
---
### Name

<code translate="no">kf schedule-job</code> - Schedule the Job for execution on a cron schedule.

### Synopsis

<pre translate="no">kf schedule-job JOB_NAME SCHEDULE [flags]</pre>

### Description

The schedule-job sub-command lets operators schedule a Job for execution on a cron schedule.

### Examples

<pre translate="no">
kf schedule-job my-job &#34;* * * * *&#34;</pre>

### Flags

<dl>
<dt><code translate="no">-h, --help</code></dt>
<dd><p>help for schedule-job</p>
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


