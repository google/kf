---
title: "kf delete-route"
weight: 100
description: "Delete a Route in the targeted Space."
---
### Name

<code translate="no">kf delete-route</code> - Delete a Route in the targeted Space.

### Synopsis

<pre translate="no">kf delete-route DOMAIN [--hostname HOSTNAME] [--path PATH] [flags]</pre>

### Examples

<pre translate="no">
# Delete the Route myapp.example.com
kf delete-route example.com --hostname myapp
# Delete a Route on a path myapp.example.com/mypath
kf delete-route example.com --hostname myapp --path /mypath
</pre>

### Flags

<dl>
<dt><code translate="no">--async</code></dt>
<dd><p>Do not wait for the action to complete on the server before returning.</p>
</dd>
<dt><code translate="no">-h, --help</code></dt>
<dd><p>help for delete-route</p>
</dd>
<dt><code translate="no">--hostname=<var translate="no">string</var></code></dt>
<dd><p>Hostname for the Route.</p>
</dd>
<dt><code translate="no">--path=<var translate="no">string</var></code></dt>
<dd><p>URL path for the Route.</p>
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


