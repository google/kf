---
title: "kf proxy-route"
weight: 100
description: "Start a local reverse proxy to a Route."
---
### Name

<code translate="no">kf proxy-route</code> - Start a local reverse proxy to a Route.

### Synopsis

<pre translate="no">kf proxy-route ROUTE [flags]</pre>

### Description

Proxy route creates a reverse HTTP proxy to the cluster's gateway on a local
port opened on the operating system's loopback device.

The proxy rewrites all HTTP requests, changing the HTTP Host header to match
the Route. If multiple Apps are mapped to the same route, the traffic sent
over the proxy will follow their routing rules with regards to weight.
If no Apps are mapped to the route, traffic sent through the proxy will
return a HTTP 404 status code.

Proxy route DOES NOT establish a direct connection to any Kubernetes resource.

For proxy to work:

* The cluster's gateway MUST be accessible from your local machine.
* The Route MUST have a public URL


### Examples

<pre translate="no">
kf proxy-route myhost.example.com</pre>

### Flags

<dl>
<dt><code translate="no">--gateway=<var translate="no">string</var></code></dt>
<dd><p>IP address of the HTTP gateway to route requests to.</p>
</dd>
<dt><code translate="no">-h, --help</code></dt>
<dd><p>help for proxy-route</p>
</dd>
<dt><code translate="no">--port=<var translate="no">int</var></code></dt>
<dd><p>Local port to listen on. (default 8080)</p>
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


