---
title: "kf bind-route-service"
weight: 100
description: "Bind a route service instance to an HTTP route."
---
### Name

<code translate="no">kf bind-route-service</code> - Bind a route service instance to an HTTP route.

### Synopsis

<pre translate="no">kf bind-route-service DOMAIN [--hostname HOSTNAME] [--path PATH] SERVICE_INSTANCE [-c PARAMETERS_AS_JSON] [flags]</pre>

### Description

PREVIEW: this feature is not ready for production use.
Binding a service to an HTTP route causes traffic to be processed by that service before the requests are forwarded to the route.


### Examples

<pre translate="no">
  kf bind-route-service company.com --hostname myapp --path mypath myauthservice</pre>

### Flags

<dl>
<dt><code translate="no">--async</code></dt>
<dd><p>Do not wait for the action to complete on the server before returning.</p>
</dd>
<dt><code translate="no">-h, --help</code></dt>
<dd><p>help for bind-route-service</p>
</dd>
<dt><code translate="no">--hostname=<var translate="no">string</var></code></dt>
<dd><p>Hostname for the Route.</p>
</dd>
<dt><code translate="no">-c, --parameters=<var translate="no">string</var></code></dt>
<dd><p>JSON object or path to a JSON file containing configuration parameters. (default &quot;{}&quot;)</p>
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


