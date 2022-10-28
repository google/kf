---
title: "kf update-user-provided-service"
weight: 100
description: "Update a standalone service instance with new credentials."
---
### Name

<code translate="no">kf update-user-provided-service</code> - Update a standalone service instance with new credentials.

### Synopsis

<pre translate="no">kf update-user-provided-service SERVICE_INSTANCE [-p CREDENTIALS] [-t TAGS] [flags]</pre>

### Description

Updates the credentials stored in the Kubernetes Secret for a user-provided service.
These credentials will be propagated to Apps.

Apps may need to be restarted to receive the updated credentials.


### Examples

<pre translate="no">
# Update an existing database service
kf update-user-provided-service db-service -p &#39;{&#34;url&#34;:&#34;mysql://...&#34;}&#39;

# Update a service with tags for autowiring
kf update-user-provided-service db-service -t &#34;mysql,database,sql&#34;
</pre>

### Flags

<dl>
<dt><code translate="no">--async</code></dt>
<dd><p>Do not wait for the action to complete on the server before returning.</p>
</dd>
<dt><code translate="no">-h, --help</code></dt>
<dd><p>help for update-user-provided-service</p>
</dd>
<dt><code translate="no">-p, --params=<var translate="no">string</var></code></dt>
<dd><p>Valid JSON object containing service-specific configuration parameters, provided in-line or in a file. (default &quot;{}&quot;)</p>
</dd>
<dt><code translate="no">-t, --tags=<var translate="no">string</var></code></dt>
<dd><p>Comma-separated tags for the service instance.</p>
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


