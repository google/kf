---
title: "kf unset-space-role"
weight: 100
description: "Unassign a Role to a Subject."
---
### Name

<code translate="no">kf unset-space-role</code> - Unassign a Role to a Subject.

### Synopsis

<pre translate="no">kf unset-space-role SUBJECT_NAME ROLE [flags]</pre>

### Examples

<pre translate="no">
# Unassign a User to a Role
kf unset-space-role john@example.com SpaceDeveloper

# Unassign a Group to a Role
kf unset-space-role my-group SpaceAuditor -t Group

# Unassign a ServiceAccount to a Role
kf unset-space-role my-sa SpaceAuditor -t ServiceAccount
</pre>

### Flags

<dl>
<dt><code translate="no">-h, --help</code></dt>
<dd><p>help for unset-space-role</p>
</dd>
<dt><code translate="no">-t, --type=<var translate="no">string</var></code></dt>
<dd><p>Type of subject, valid values are Group|ServiceAccount|User(default). (default &quot;User&quot;)</p>
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


