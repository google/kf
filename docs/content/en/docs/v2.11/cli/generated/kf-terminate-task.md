---
title: "kf terminate-task"
weight: 100
description: "Terminate a running Task."
---
### Name

<code translate="no">kf terminate-task</code> - Terminate a running Task.

### Synopsis

<pre translate="no">kf terminate-task {TASK_NAME | APP_NAME TASK_ID} [flags]</pre>

### Description

Allows operators to terminate a running Task on an App.

### Examples

<pre translate="no">
# Terminate Task by Task name
kf terminate-task my-task-name

# Terminate Task by App name and Task ID
kf terminate-task my-app 1
</pre>

### Flags

<dl>
<dt><code translate="no">--async</code></dt>
<dd><p>Do not wait for the action to complete on the server before returning.</p>
</dd>
<dt><code translate="no">-h, --help</code></dt>
<dd><p>help for terminate-task</p>
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


