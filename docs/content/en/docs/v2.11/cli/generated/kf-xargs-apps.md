---
title: "kf xargs-apps"
weight: 100
description: "Run a command for every App."
---
### Name

<code translate="no">kf xargs-apps</code> - Run a command for every App.

### Synopsis

<pre translate="no">kf xargs-apps [flags]</pre>

### Description

Run a command for every App in targeted spaces.

### Examples

<pre translate="no">
# Example: restart all apps in all spaces
kf xargs-apps --all-namespaces -- kf restart {{.Name}} --space {{.Space}}

# Example: restage all apps in all spaces
kf xargs-apps --all-namespaces -- kf restage {{.Name}} --space {{.Space}}

# Example: stop all apps in spaces &#39;space1&#39; and &#39;space2&#39;
kf xargs-apps --space space1,space2 -- kf stop {{.Name}} --space {{.Space}}

# Example: use kubectl to label all apps in the default space
kf xargs-apps -- kubectl label apps -n {{.Space}} {{.Name}} environment=prod</pre>

### Flags

<dl>
<dt><code translate="no">--all-namespaces</code></dt>
<dd><p>Enables targeting all spaces in the cluster.</p>
</dd>
<dt><code translate="no">--dry-run</code></dt>
<dd><p>Enables dry-run mode, commands are printed but will not be executed. (default true)</p>
</dd>
<dt><code translate="no">-h, --help</code></dt>
<dd><p>help for xargs-apps</p>
</dd>
<dt><code translate="no">--resource-concurrency=<var translate="no">int</var></code></dt>
<dd><p>Number of apps within a space that may be operated on in parallel. Total concurrency will be upto space-concurrency * app-concurrency. -1 for no limit. (default 1)</p>
</dd>
<dt><code translate="no">--space-concurrency=<var translate="no">int</var></code></dt>
<dd><p>Number of spaces that may be operated on in parallel. -1 for no limit. (default -1)</p>
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


