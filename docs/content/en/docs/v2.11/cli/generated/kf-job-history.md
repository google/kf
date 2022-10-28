---
title: "kf job-history"
weight: 100
description: "List the execution history of a Job."
---
### Name

<code translate="no">kf job-history</code> - List the execution history of a Job.

### Synopsis

<pre translate="no">kf job-history JOB_NAME [flags]</pre>

### Description

The job-history sub-command lets operators view the execution history of a Job.

### Examples

<pre translate="no">
kf job-history my-job</pre>

### Flags

<dl>
<dt><code translate="no">--allow-missing-template-keys</code></dt>
<dd><p>If true, ignore any errors in templates when a field or map key is missing in the template. Only applies to golang and jsonpath output formats. (default true)</p>
</dd>
<dt><code translate="no">-h, --help</code></dt>
<dd><p>help for job-history</p>
</dd>
<dt><code translate="no">-o, --output=<var translate="no">string</var></code></dt>
<dd><p>Output format. One of: go-template|go-template-file|json|jsonpath|jsonpath-as-json|jsonpath-file|name|template|templatefile|yaml.</p>
</dd>
<dt><code translate="no">--show-managed-fields</code></dt>
<dd><p>If true, keep the managedFields when printing objects in JSON or YAML format.</p>
</dd>
<dt><code translate="no">--template=<var translate="no">string</var></code></dt>
<dd><p>Template string or path to template file to use when -o=go-template, -o=go-template-file. The template format is <a href="http://golang.org/pkg/text/template/#pkg-overview">golang templates</a>.</p>
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


