---
title: "kf fix-orphaned-bindings"
weight: 100
description: "Fix bindings without an app owner in a space."
---
### Name

<code translate="no">kf fix-orphaned-bindings</code> - Fix bindings without an app owner in a space.

### Synopsis

<pre translate="no">kf fix-orphaned-bindings [flags]</pre>

### Description

Fix bindings with a missing owner reference to an App.

### Examples

<pre translate="no">
# Identify broken bindings in the targeted space.
kf fix-orphaned-bindings

# Fix bindings in the targeted space.
kf fix-orphaned-bindings --dry-run=false
</pre>

### Flags

<dl>
<dt><code translate="no">--dry-run</code></dt>
<dd><p>Run the command without applying changes. (default true)</p>
</dd>
<dt><code translate="no">-h, --help</code></dt>
<dd><p>help for fix-orphaned-bindings</p>
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


