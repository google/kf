---
title: "kf ssh"
weight: 100
description: "Open a shell on an App instance."
---
### Name

<code translate="no">kf ssh</code> - Open a shell on an App instance.

### Synopsis

<pre translate="no">kf ssh APP_NAME [flags]</pre>

### Description

Opens a shell on an App instance using the Pod exec endpoint.

This command mimics CF's SSH command by opening a connection to the
Kubernetes control plane which spawns a process in a Pod.

The command connects to an arbitrary Pod that matches the App's runtime
labels. If you want a specific Pod, use the pod/<podname> notation.

NOTE: Traffic is encrypted between the CLI and the control plane, and
between the control plane and Pod. A malicious Kubernetes control plane
could observe the traffic.


### Examples

<pre translate="no">
# Open a shell to a specific App
kf ssh myapp

# Open a shell to a specific Pod
kf ssh pod/myapp-revhex-podhex

# Start a different command with args
kf ssh myapp -c /my/command -c arg1 -c arg2
</pre>

### Flags

<dl>
<dt><code translate="no">-c, --command=<var translate="no">stringArray</var></code></dt>
<dd><p>Command to run for the shell. Subsequent definitions will be used as args. (default [/bin/bash])</p>
</dd>
<dt><code translate="no">--container=<var translate="no">string</var></code></dt>
<dd><p>Container to start the command in. (default &quot;user-container&quot;)</p>
</dd>
<dt><code translate="no">-T, --disable-pseudo-tty</code></dt>
<dd><p>Don't use a TTY when executing.</p>
</dd>
<dt><code translate="no">-h, --help</code></dt>
<dd><p>help for ssh</p>
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


