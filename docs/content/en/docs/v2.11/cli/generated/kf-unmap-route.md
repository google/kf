---
title: "kf unmap-route"
weight: 100
description: "Revoke an App's access to receive traffic from the Route."
---
### Name

<code translate="no">kf unmap-route</code> - Revoke an App's access to receive traffic from the Route.

### Synopsis

<pre translate="no">kf unmap-route APP_NAME DOMAIN [--hostname HOSTNAME] [--path PATH] [flags]</pre>

### Description

Unmapping an App from a Route will cause traffic matching the Route to no
longer be forwarded to the App.

The App may still receive traffic from an unmapped Route for a small period
of time while the traffic rules on the gateways are propagated.

The Route will re-balance its routing weights so other Apps mapped to it
will receive the traffic. If no other Apps are bound the Route will return
a 404 HTTP status code.


### Examples

<pre translate="no">
# Unmap myapp.example.com from myapp in the targeted Space
kf unmap-route myapp example.com --hostname myapp

# Unmap the Route in a specific Space
kf unmap-route --space myspace myapp example.com --hostname myapp

# Unmap a Route with a path
kf unmap-route myapp example.com --hostname myapp --path /mypath
</pre>

### Flags

<dl>
<dt><code translate="no">--async</code></dt>
<dd><p>Do not wait for the action to complete on the server before returning.</p>
</dd>
<dt><code translate="no">--destination-port=<var translate="no">int32</var></code></dt>
<dd><p>Port on the App the Route will connect to.</p>
</dd>
<dt><code translate="no">-h, --help</code></dt>
<dd><p>help for unmap-route</p>
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


