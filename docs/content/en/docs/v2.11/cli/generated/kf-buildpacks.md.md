---
title: "kf buildpacks"
weight: 100
description: "List buildpacks in the targeted Space."
---
### Name

<code translate="no">kf buildpacks</code> - List buildpacks in the targeted Space.

### Synopsis

<pre translate="no">kf buildpacks [flags]</pre>

### Description

List the buildpacks available in the Space to Apps being built with
buildpacks.

The buildpacks available to an App depend on the Stack it uses.
To ensure reproducibility in Builds, Apps should explicitly declare the
Stack they use.


### Examples

<pre translate="no">
kf buildpacks</pre>

### Flags

<dl>
<dt><code translate="no">-h, --help</code></dt>
<dd><p>help for buildpacks</p>
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


