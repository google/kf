---
title: "kf stacks"
weight: 100
description: "List stacks in the targeted Space."
---
### Name

<code translate="no">kf stacks</code> - List stacks in the targeted Space.

### Synopsis

<pre translate="no">kf stacks [flags]</pre>

### Description

Stacks contain information about how to build and run an App.
Each stack contains:

*  A unique name to identify it.
*  A build image, the image used to build the App, this usually contains
	 things like compilers, libraries, sources and build frameworks.
*  A run image, the image App instances will run within. These images
	 are usually lightweight and contain just enough to run an App.
*  A list of applicable buildpacks available via the bulidpacks command.


### Examples

<pre translate="no">
kf stacks</pre>

### Flags

<dl>
<dt><code translate="no">-h, --help</code></dt>
<dd><p>help for stacks</p>
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


