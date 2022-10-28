---
title: "kf create-user-provided-service"
weight: 100
description: "Create a standalone service instance from existing credentials."
---
### Name

<code translate="no">kf create-user-provided-service</code> - Create a standalone service instance from existing credentials.

### Synopsis

<pre translate="no">kf create-user-provided-service SERVICE_INSTANCE [-p CREDENTIALS] [-t TAGS] [flags]</pre>

### Description

Creates a standalone service instance from existing credentials.
User-provided services can be used to inject credentials for services managed
outside of Kf into Apps.

Credentials are stored in a Kubernetes Secret in the Space the service is
created in. On GKE these Secrets are encrypted at rest and can optionally
be encrypted using KMS.


### Examples

<pre translate="no">
# Bring an existing database service
kf create-user-provided-service db-service -p &#39;{&#34;url&#34;:&#34;mysql://...&#34;}&#39;

# Create a service with tags for autowiring
kf create-user-provided-service db-service -t &#34;mysql,database,sql&#34;
</pre>

### Flags

<dl>
<dt><code translate="no">--async</code></dt>
<dd><p>Do not wait for the action to complete on the server before returning.</p>
</dd>
<dt><code translate="no">-h, --help</code></dt>
<dd><p>help for create-user-provided-service</p>
</dd>
<dt><code translate="no">--mock-class=<var translate="no">string</var></code></dt>
<dd><p>Mock class name to use in VCAP_SERVICES rather than 'user-provided'.</p>
</dd>
<dt><code translate="no">--mock-plan=<var translate="no">string</var></code></dt>
<dd><p>Mock plan name to use in VCAP_SERVICES rather than blank.</p>
</dd>
<dt><code translate="no">-p, --parameters=<var translate="no">string</var></code></dt>
<dd><p>JSON object or path to a JSON file containing configuration parameters. (default &quot;{}&quot;)</p>
</dd>
<dt><code translate="no">--params=<var translate="no">string</var></code></dt>
<dd><p>JSON object or path to a JSON file containing configuration parameters. DEPRECATED: use --parameters instead. (default &quot;{}&quot;)</p>
</dd>
<dt><code translate="no">-r, --route=<var translate="no">string</var></code></dt>
<dd><p>URL to which requests for bound routes will be forwarded. Scheme must be https. NOTE: This is a preivew feature.</p>
</dd>
<dt><code translate="no">-t, --tags=<var translate="no">string</var></code></dt>
<dd><p>User-defined tags to differentiate services during injection.</p>
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


