---
title: "kf env"
weight: 100
description: "Print information about an App's environment variables."
---
### Name

<code translate="no">kf env</code> - Print information about an App's environment variables.

### Synopsis

<pre translate="no">kf env APP_NAME [flags]</pre>

### Description

The env command gets the names and values of developer managed
environment variables for an App.

Environment variables are evaluated in the following order with later values
overriding earlier ones with the same name:

1. Space (set by administrators)
1. App (set by developers)
1. System (set by Kf)

Environment variables containing variable substitution "$(...)" are
replaced at runtime by Kubernetes.
Kf provides the following runtime environment variables:

* CF_INSTANCE_ADDR: The cluster-visible IP:PORT of the App instance.
* CF_INSTANCE_GUID: The UUID of the App instance.
* INSTANCE_GUID: Alias of CF_INSTANCE_GUID
* CF_INSTANCE_INDEX: The index number of the App instance, this will ALWAYS be 0.
* INSTANCE_INDEX: Alias of CF_INSTANCE_INDEX
* CF_INSTANCE_IP: The cluster-visible IP of the App instance.
* CF_INSTANCE_INTERNAL_IP: Alias of CF_INSTANCE_IP
* VCAP_APP_HOST: Alias of CF_INSTANCE_IP
* CF_INSTANCE_PORT: The cluster-visible port of the App instance. In Kf this is the same as PORT.
* DATABASE_URL: The first URI found in a VCAP_SERVICES credential.
* DISK_LIMIT: The maximum amount of disk storage in MB the App can use.
* LANG: Required by buildpacks to ensure consistent script load order.
* MEMORY_LIMIT: The maximum amount of memory in MB the App can consume.
* PORT: The port the App should listen on for requests.
* VCAP_APP_PORT: Alias of PORT
* VCAP_APPLICATION: A JSON structure containing app metadata.
* VCAP_SERVICES: A JSON structure specifying bound services.


### Examples

<pre translate="no">
kf env myapp</pre>

### Flags

<dl>
<dt><code translate="no">-h, --help</code></dt>
<dd><p>help for env</p>
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


