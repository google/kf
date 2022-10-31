---
title: "kf proxy"
weight: 100
description: "Start a local reverse proxy to an App."
---
### Name

<code translate="no">kf proxy</code> - Start a local reverse proxy to an App.

### Synopsis

<pre translate="no">kf proxy APP_NAME [flags]</pre>

### Description

Proxy creates a reverse HTTP proxy to the cluster's gateway on a local
port opened on the operating system's loopback device.

The proxy rewrites all HTTP requests, changing the HTTP Host header
and adding an additional header X-Kf-App to ensure traffic reaches
the specified App even if multiple are attached to the same route.

Proxy does not establish a direct connection to the App.

For proxy to work:

* The cluster's gateway must be accessible from your local machine.
* The App must have a public URL

If you need to establish a direct connection to an App, use the
port-forward command in kubectl. It establishes a proxied connection
directly to a port on a pod via the Kubernetes cluster. port-forward
bypasses all routing.


### Examples

<pre translate="no">
kf proxy myapp</pre>

### Flags

<dl>
<dt><code translate="no">--gateway=<var translate="no">string</var></code></dt>
<dd><p>IP address of the HTTP gateway to route requests to.</p>
</dd>
<dt><code translate="no">-h, --help</code></dt>
<dd><p>help for proxy</p>
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


