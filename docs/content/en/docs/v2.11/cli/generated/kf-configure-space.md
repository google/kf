---
title: "kf configure-space"
weight: 100
description: "Set configuration for a Space."
---
### Name

<code translate="no">kf configure-space</code> - Set configuration for a Space.

### Synopsis

<pre translate="no">kf configure-space [subcommand] [flags]</pre>

### Description

The configure-space sub-command allows operators to configure
individual fields on a Space.

In Kf, most configuration can be overridden at the Space level.

NOTE: The Space is reconciled every time changes are made using this command.
If you want to configure Spaces in automation it's better to use kubectl.


### Flags

<dl>
<dt><code translate="no">-h, --help</code></dt>
<dd><p>help for configure-space</p>
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


