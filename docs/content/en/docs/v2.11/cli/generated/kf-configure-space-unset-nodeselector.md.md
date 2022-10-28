---
title: "kf configure-space unset-nodeselector"
weight: 100
description: "Unset a Space wide node selector for all Apps."
---
### Name

<code translate="no">kf configure-space unset-nodeselector</code> - Unset a Space wide node selector for all Apps.

### Synopsis

<pre translate="no">kf configure-space unset-nodeselector [SPACE_NAME] NS_VAR_NAME [flags]</pre>

### Examples

<pre translate="no">
# Configure the Space &#34;my-space&#34;
kf configure-space unset-nodeselector my-space DiskType
# Configure the targeted Space
kf configure-space unset-nodeselector DiskType
</pre>

### Flags

<dl>
<dt><code translate="no">--async</code></dt>
<dd><p>Do not wait for the action to complete on the server before returning.</p>
</dd>
<dt><code translate="no">-h, --help</code></dt>
<dd><p>help for unset-nodeselector</p>
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


