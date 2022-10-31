---
title: "kf create-service"
weight: 100
description: "Create a service instance from a marketplace template."
---
### Name

<code translate="no">kf create-service</code> - Create a service instance from a marketplace template.

### Synopsis

<pre translate="no">kf create-service SERVICE PLAN SERVICE_INSTANCE [-c PARAMETERS_AS_JSON] [-b service-broker] [-t TAGS] [flags]</pre>

### Description

Create service creates a new ServiceInstance using a template from the
marketplace.


### Examples

<pre translate="no">
# Creates a new instance of a db-service with the name mydb, plan silver, and provisioning configuration
kf create-service db-service silver mydb -c &#39;{&#34;ram_gb&#34;:4}&#39;

# Creates a new instance of a db-service from the broker named local-broker
kf create-service db-service silver mydb -c ~/workspace/tmp/instance_config.json -b local-broker

# Creates a new instance of a db-service with the name mydb and override tags
kf create-service db-service silver mydb -t &#34;list, of, tags&#34;</pre>

### Flags

<dl>
<dt><code translate="no">--async</code></dt>
<dd><p>Do not wait for the action to complete on the server before returning.</p>
</dd>
<dt><code translate="no">-b, --broker=<var translate="no">string</var></code></dt>
<dd><p>Name of the service broker that will create the instance.</p>
</dd>
<dt><code translate="no">-h, --help</code></dt>
<dd><p>help for create-service</p>
</dd>
<dt><code translate="no">-c, --parameters=<var translate="no">string</var></code></dt>
<dd><p>JSON object or path to a JSON file containing configuration parameters. (default &quot;{}&quot;)</p>
</dd>
<dt><code translate="no">-t, --tags=<var translate="no">string</var></code></dt>
<dd><p>User-defined tags to differentiate services during injection.</p>
</dd>
<dt><code translate="no">--timeout=<var translate="no">duration</var></code></dt>
<dd><p>Amount of time to wait for the operation to complete. Valid units are &quot;s&quot;, &quot;m&quot;, &quot;h&quot;. (default 30m0s)</p>
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


