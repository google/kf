---
title: "kf enable-autoscaling"
weight: 100
description: "Enable autoscaling for App."
---
### Name

<code translate="no">kf enable-autoscaling</code> - Enable autoscaling for App.

### Synopsis

<pre translate="no">kf enable-autoscaling APP_NAME [flags]</pre>

### Description

Enabling autoscaling creates a HorizontalPodAutoscaler (HPA) for an App. HPA
controls the number of replicas the App should have based on CPU utilization.

Autoscaler will only take effect after autoscaling limits are set and autoscaling
rules are added.

Set autoscaling limits by running:
	kf update-autoscaling-limits APP_NAME MIN_INSTANCES MAX_INSTANCES

Add rules by running:
	kf create-autoscaling-rule APP_NAME RULE_TYPE MIN_THRESHOLD MAX_THRESHOLD


### Examples

<pre translate="no">
kf enable-autoscaling myapp</pre>

### Flags

<dl>
<dt><code translate="no">--async</code></dt>
<dd><p>Do not wait for the action to complete on the server before returning.</p>
</dd>
<dt><code translate="no">-h, --help</code></dt>
<dd><p>help for enable-autoscaling</p>
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


