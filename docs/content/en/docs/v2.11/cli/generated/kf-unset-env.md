---
title: "kf unset-env"
weight: 100
description: "Delete an environment variable on an App."
---
### Name

<code translate="no">kf unset-env</code> - Delete an environment variable on an App.

### Synopsis

<pre translate="no">kf unset-env APP_NAME ENV_VAR_NAME [flags]</pre>

### Description

Removes an environment variable by name from an App.
Any existing environment variable(s) on the App with the same name will be removed.

Environment variables that are set by Kf or on the Space will be unaffected.

Apps will be updated without downtime.


### Examples

<pre translate="no">
kf unset-env myapp FOO</pre>

### Flags

<dl>
<dt><code translate="no">--async</code></dt>
<dd><p>Do not wait for the action to complete on the server before returning.</p>
</dd>
<dt><code translate="no">-h, --help</code></dt>
<dd><p>help for unset-env</p>
</dd>
<dt><code translate="no">--no-short-circuit-wait</code></dt>
<dd><p>Allow the CLI to skip waiting if the mutation does not impact a running resource.</p>
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


