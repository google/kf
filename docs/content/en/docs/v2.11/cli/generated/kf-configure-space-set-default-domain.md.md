---
title: "kf configure-space set-default-domain"
weight: 100
description: "Set or create a default domain for a Space."
---
### Name

<code translate="no">kf configure-space set-default-domain</code> - Set or create a default domain for a Space.

### Synopsis

<pre translate="no">kf configure-space set-default-domain [SPACE_NAME] DOMAIN [flags]</pre>

### Examples

<pre translate="no">
# Configure the Space &#34;my-space&#34;
kf configure-space set-default-domain my-space myspace.mycompany.com
# Configure the targeted Space
kf configure-space set-default-domain myspace.mycompany.com
</pre>

### Flags

<dl>
<dt><code translate="no">--async</code></dt>
<dd><p>Do not wait for the action to complete on the server before returning.</p>
</dd>
<dt><code translate="no">-h, --help</code></dt>
<dd><p>help for set-default-domain</p>
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


