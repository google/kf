---
title: "kf set-env"
weight: 100
description: "Create or update an environment variable for an App."
---
### Name

<code translate="no">kf set-env</code> - Create or update an environment variable for an App.

### Synopsis

<pre translate="no">kf set-env APP_NAME ENV_VAR_NAME ENV_VAR_VALUE [flags]</pre>

### Description

Sets an environment variable for an App. Existing environment
variable(s) on the App with the same name will be replaced.

Apps will be updated without downtime.


### Examples

<pre translate="no">
# Set an environment variable on an App.
kf set-env myapp ENV production

# Don&#39;t wait for the App to restart.
kf set-env --async myapp ENV production

# Set an environment variable that starts with a dash.
kf set-env myapp -- JAVA_OPTS -Dtest=sometest
</pre>

### Flags

<dl>
<dt><code translate="no">--async</code></dt>
<dd><p>Do not wait for the action to complete on the server before returning.</p>
</dd>
<dt><code translate="no">-h, --help</code></dt>
<dd><p>help for set-env</p>
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


