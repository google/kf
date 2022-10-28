---
title: "kf create-route"
weight: 100
description: "Create a traffic routing rule for a host+path pair."
---
### Name

<code translate="no">kf create-route</code> - Create a traffic routing rule for a host+path pair.

### Synopsis

<pre translate="no">kf create-route DOMAIN [--hostname HOSTNAME] [--path PATH] [flags]</pre>

### Description

Creating a Route allows Apps to declare they want to receive traffic on
the same host/domain/path combination.

Routes without any bound Apps (or with only stopped Apps) will return a 404
HTTP status code.

Kf doesn't enforce Route uniqueness between Spaces. It's recommended
to provide each Space with its own subdomain instead.


### Examples

<pre translate="no">
kf create-route example.com --hostname myapp # myapp.example.com
kf create-route --space myspace example.com --hostname myapp # myapp.example.com
kf create-route example.com --hostname myapp --path /mypath # myapp.example.com/mypath
kf create-route --space myspace myapp.example.com # myapp.example.com

# Using SPACE to match &#39;cf&#39;
kf create-route myspace example.com --hostname myapp # myapp.example.com
kf create-route myspace example.com --hostname myapp --path /mypath # myapp.example.com/mypath
</pre>

### Flags

<dl>
<dt><code translate="no">--async</code></dt>
<dd><p>Do not wait for the action to complete on the server before returning.</p>
</dd>
<dt><code translate="no">-h, --help</code></dt>
<dd><p>help for create-route</p>
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


