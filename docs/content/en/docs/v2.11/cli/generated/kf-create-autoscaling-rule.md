---
title: "kf create-autoscaling-rule"
weight: 100
description: "Create autoscaling rule for App."
---
### Name

<code translate="no">kf create-autoscaling-rule</code> - Create autoscaling rule for App.

### Synopsis

<pre translate="no">kf create-autoscaling-rule APP RULE_TYPE MIN_THRESHOLD MAX_THRESHOLD [flags]</pre>

### Description

Create an autoscaling rule for App.

The only supported rule type is CPU. It is the target
percentage. It is calculated by taking the average of MIN_THRESHOLD
and MAX_THRESHOLD.

The range of MIN_THRESHOLD and MAX_THRESHOLD is 1 to 100 (percent).


### Examples

<pre translate="no">
# Scale myapp based on CPU load targeting 50% utilization (halfway between 20 and 80)
kf create-autoscaling-rule myapp CPU 20 80
</pre>

### Flags

<dl>
<dt><code translate="no">--async</code></dt>
<dd><p>Do not wait for the action to complete on the server before returning.</p>
</dd>
<dt><code translate="no">-h, --help</code></dt>
<dd><p>help for create-autoscaling-rule</p>
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


