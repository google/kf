---
title: "kf create-space"
weight: 100
description: "Create a Space with the given name."
---
### Name

<code translate="no">kf create-space</code> - Create a Space with the given name.

### Synopsis

<pre translate="no">kf create-space NAME [flags]</pre>

### Examples

<pre translate="no">
# Create a Space with custom domains.
kf create-space my-space --domain my-space.my-company.com

# Create a Space that uses unique storage and service accounts.
kf create-space my-space --container-registry gcr.io/my-project --build-service-account myserviceaccount

# Set running and staging environment variables for Apps and Builds.
kf create-space my-space --run-env=ENVIRONMENT=nonprod --stage-env=ENVIRONMENT=nonprod,JDK_VERSION=8
</pre>

### Flags

<dl>
<dt><code translate="no">--build-service-account=<var translate="no">string</var></code></dt>
<dd><p>Service account that Builds will use.</p>
</dd>
<dt><code translate="no">--container-registry=<var translate="no">string</var></code></dt>
<dd><p>Container registry built Apps and source code will be stored in.</p>
</dd>
<dt><code translate="no">--domain=<var translate="no">stringArray</var></code></dt>
<dd><p>Sets the valid domains for the Space. The first provided domain is the default.</p>
</dd>
<dt><code translate="no">-h, --help</code></dt>
<dd><p>help for create-space</p>
</dd>
<dt><code translate="no">--run-env=<var translate="no">stringToString</var></code></dt>
<dd><p>Sets the running environment variables for all Apps in the Space.</p>
</dd>
<dt><code translate="no">--stage-env=<var translate="no">stringToString</var></code></dt>
<dd><p>Sets the staging environment variables for all Builds in the Space.</p>
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


