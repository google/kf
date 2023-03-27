---
title: "kf map-route"
weight: 100
description: "Grant an App access to receive traffic from the Route."
---
### Name

<code translate="no">kf map-route</code> - Grant an App access to receive traffic from the Route.

### Synopsis

<pre translate="no">kf map-route APP_NAME DOMAIN [--hostname HOSTNAME] [--path PATH] [--weight WEIGHT] [flags]</pre>

### Description

Mapping an App to a Route will cause traffic to be forwarded to the App if
the App has instances that are running and healthy.

If multiple Apps are mapped to the same Route they will split traffic
between them roughly evenly. Incoming network traffic is handled by multiple
gateways which update their routing tables with slight delays and route
independently. Because of this, traffic routing may not appear even but it
will converge over time.


### Examples

<pre translate="no">
kf map-route myapp example.com --hostname myapp # myapp.example.com
kf map-route myapp myapp.example.com # myapp.example.com
kf map-route myapp example.com --hostname myapp --weight 2 # myapp.example.com, myapp receives 2x traffic
kf map-route --space myspace myapp example.com --hostname myapp # myapp.example.com
kf map-route myapp example.com --hostname myapp --path /mypath # myapp.example.com/mypath
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
<dd><p>help for map-route</p>
</dd>
<dt><code translate="no">--hostname=<var translate="no">string</var></code></dt>
<dd><p>Hostname for the Route.</p>
</dd>
<dt><code translate="no">--no-short-circuit-wait</code></dt>
<dd><p>Allow the CLI to skip waiting if the mutation does not impact a running resource.</p>
</dd>
<dt><code translate="no">--path=<var translate="no">string</var></code></dt>
<dd><p>URL path for the Route.</p>
</dd>
<dt><code translate="no">--weight=<var translate="no">int32</var></code></dt>
<dd><p>Weight for the Route. (default 1)</p>
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


