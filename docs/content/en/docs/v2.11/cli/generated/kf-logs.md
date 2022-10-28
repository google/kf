---
title: "kf logs"
weight: 100
description: "Show logs for an App."
---
### Name

<code translate="no">kf logs</code> - Show logs for an App.

### Synopsis

<pre translate="no">kf logs APP_NAME [flags]</pre>

### Description

Logs are streamed from the Kubernetes log endpoint for each running
App instance.

If App instances change or the connection to Kubernetes times out the
log stream may show duplicate logs.

Logs are retained for App instances as space permits on the cluster,
but will be deleted if space is low or past their retention date.
Cloud Logging is a more reliable mechanism to access historical logs.

If you need logs for a particular instance use the <code>kubectl</code> CLI.


### Examples

<pre translate="no">
# Follow/tail the log stream
kf logs myapp

# Follow/tail the log stream with 20 lines of context
kf logs myapp -n 20

# Get recent logs from the App
kf logs myapp --recent

# Get the most recent 200 lines of logs from the App
kf logs myapp --recent -n 200

# Get the logs of Tasks running from the App
kf logs myapp --task
</pre>

### Flags

<dl>
<dt><code translate="no">-h, --help</code></dt>
<dd><p>help for logs</p>
</dd>
<dt><code translate="no">-n, --number=<var translate="no">int</var></code></dt>
<dd><p>Show the last N lines of logs. (default 10)</p>
</dd>
<dt><code translate="no">--recent</code></dt>
<dd><p>Dump recent logs instead of tailing.</p>
</dd>
<dt><code translate="no">--task</code></dt>
<dd><p>Tail Task logs instead of App.</p>
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


