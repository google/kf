---
title: "kf wrap-v2-buildpack"
weight: 100
description: "Create a V3 buildpack that wraps a V2 buildpack."
---
### Name

<code translate="no">kf wrap-v2-buildpack</code> - Create a V3 buildpack that wraps a V2 buildpack.

### Synopsis

<pre translate="no">kf wrap-v2-buildpack NAME V2_BUILDPACK_URL|PATH [flags]</pre>

### Description

Creates a V3 buildpack that wraps a V2 buildpack.

The resulting buildpack can then be used with other V3 buildpacks by
creating a builder. See
https://buildpacks.io/docs/operator-guide/create-a-builder/ for more
information.

A V3 buildpack is packaged as an OCI container. If the --publish flag
is provided, then the container will be published to the corresponding
container repository.

This command uses other CLIs under the hood. This means the following
CLIs need to be available on the path:
* go
* git
* pack
* cp
* unzip

We recommend using Cloud Shell to ensure these CLIs are available and
the correct version.


### Examples

<pre translate="no">
# Download buildpack from the given git URL. Uses the git CLI to
# download the repository.
kf wrap-v2-buildpack gcr.io/some-project/some-name https://github.com/some/buildpack

# Creates the buildpack from the given path.
kf wrap-v2-buildpack gcr.io/some-project/some-name path/to/buildpack

# Creates the buildpack from the given archive file.
kf wrap-v2-buildpack gcr.io/some-project/some-name path/to/buildpack.zip
</pre>

### Flags

<dl>
<dt><code translate="no">--builder-repo=<var translate="no">string</var></code></dt>
<dd><p>Builder repo to use. (default &quot;github.com/poy/buildpackapplifecycle/builder&quot;)</p>
</dd>
<dt><code translate="no">--buildpack-stacks=<var translate="no">stringArray</var></code></dt>
<dd><p>Stack(s) this buildpack will be compatible with. (default [google])</p>
</dd>
<dt><code translate="no">--buildpack-version=<var translate="no">string</var></code></dt>
<dd><p>Version of the resulting buildpack. This will be used as the image tag. (default &quot;0.0.1&quot;)</p>
</dd>
<dt><code translate="no">-h, --help</code></dt>
<dd><p>help for wrap-v2-buildpack</p>
</dd>
<dt><code translate="no">--launcher-repo=<var translate="no">string</var></code></dt>
<dd><p>Launcher repo to use. (default &quot;github.com/poy/buildpackapplifecycle/launcher&quot;)</p>
</dd>
<dt><code translate="no">--output-dir=<var translate="no">string</var></code></dt>
<dd><p>Output directory for the buildpack data (before it's packed). If left empty, a temp dir will be used.</p>
</dd>
<dt><code translate="no">--publish</code></dt>
<dd><p>Publish the buildpack image.</p>
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


