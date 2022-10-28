---
title: "kf push"
weight: 100
description: "Create a new App or apply updates to an existing one."
---
### Name

<code translate="no">kf push</code> - Create a new App or apply updates to an existing one.

### Synopsis

<pre translate="no">kf push APP_NAME [flags]</pre>

### Examples

<pre translate="no">
kf push myapp
kf push myapp --buildpack my.special.buildpack # Discover via kf buildpacks
kf push myapp --env FOO=bar --env BAZ=foo
kf push myapp --stack cloudfoundry/cflinuxfs3 # Use a cflinuxfs3 runtime
kf push myapp --health-check-http-endpoint /myhealthcheck # Specify a healthCheck for the app
</pre>

### Flags

<dl>
<dt><code translate="no">--app-suffix=<var translate="no">string</var></code></dt>
<dd><p>Suffix to append to the end of every pushed App.</p>
</dd>
<dt><code translate="no">--args=<var translate="no">stringArray</var></code></dt>
<dd><p>Override the args for the image. Can't be used with the command flag.</p>
</dd>
<dt><code translate="no">-b, --buildpack=<var translate="no">string</var></code></dt>
<dd><p>Use the specified buildpack rather than the built-in.</p>
</dd>
<dt><code translate="no">-c, --command=<var translate="no">string</var></code></dt>
<dd><p>Startup command for the App, this overrides the default command specified by the web process.</p>
</dd>
<dt><code translate="no">--container-registry=<var translate="no">string</var></code></dt>
<dd><p>Container registry to push images to.</p>
</dd>
<dt><code translate="no">--cpu-cores=<var translate="no">string</var></code></dt>
<dd><p>Number of dedicated CPU cores to give each App instance (for example 100m, 0.5, 1, 2). For more information see https://kubernetes.io/docs/tasks/configure-pod-container/assign-cpu-resource/.</p>
</dd>
<dt><code translate="no">-k, --disk-quota=<var translate="no">string</var></code></dt>
<dd><p>Size of dedicated ephemeral disk attached to each App instance (for example 512M, 2G, 1T).</p>
</dd>
<dt><code translate="no">--docker-image=<var translate="no">string</var></code></dt>
<dd><p>Docker image to deploy rather than building from source.</p>
</dd>
<dt><code translate="no">--dockerfile=<var translate="no">string</var></code></dt>
<dd><p>Path to the Dockerfile to build. Relative to the source root.</p>
</dd>
<dt><code translate="no">--entrypoint=<var translate="no">string</var></code></dt>
<dd><p>Overwrite the default entrypoint of the image. Can't be used with the command flag.</p>
</dd>
<dt><code translate="no">-e, --env=<var translate="no">stringArray</var></code></dt>
<dd><p>Set environment variables. Multiple can be set by using the flag multiple times (for example, NAME=VALUE).</p>
</dd>
<dt><code translate="no">--health-check-http-endpoint=<var translate="no">string</var></code></dt>
<dd><p>HTTP endpoint to target as part of the health-check. Only valid if health-check-type is http.</p>
</dd>
<dt><code translate="no">-u, --health-check-type=<var translate="no">string</var></code></dt>
<dd><p>App health check type: http, port (default) or process.</p>
</dd>
<dt><code translate="no">-h, --help</code></dt>
<dd><p>help for push</p>
</dd>
<dt><code translate="no">-i, --instances=<var translate="no">int32</var></code></dt>
<dd><p>If set, overrides the number of instances of the App to run, -1 represents non-user input. (default -1)</p>
</dd>
<dt><code translate="no">-f, --manifest=<var translate="no">string</var></code></dt>
<dd><p>Path to the application manifest.</p>
</dd>
<dt><code translate="no">-m, --memory-limit=<var translate="no">string</var></code></dt>
<dd><p>Amount of dedicated RAM to give each App instance (for example 512M, 6G, 1T).</p>
</dd>
<dt><code translate="no">--no-manifest</code></dt>
<dd><p>Do not read the manifest file even if one exists.</p>
</dd>
<dt><code translate="no">--no-route</code></dt>
<dd><p>Prevents the App from being reachable once deployed.</p>
</dd>
<dt><code translate="no">--no-start</code></dt>
<dd><p>Build but do not run the App.</p>
</dd>
<dt><code translate="no">-p, --path=<var translate="no">string</var></code></dt>
<dd><p>If specified, overrides the path to the source code.</p>
</dd>
<dt><code translate="no">--random-route</code></dt>
<dd><p>Create a random Route for this App if it doesn't have one.</p>
</dd>
<dt><code translate="no">--route=<var translate="no">stringArray</var></code></dt>
<dd><p>Use the routes flag to provide multiple HTTP and TCP routes. Each Route for this App is created if it does not already exist.</p>
</dd>
<dt><code translate="no">-s, --stack=<var translate="no">string</var></code></dt>
<dd><p>Base image to use for to use for Apps created with a buildpack.</p>
</dd>
<dt><code translate="no">--task</code></dt>
<dd><p>Push an App to execute Tasks only. The App will be built, but not run. It will not have a route assigned.</p>
</dd>
<dt><code translate="no">-t, --timeout=<var translate="no">int</var></code></dt>
<dd><p>Amount of time the App can be unhealthy before declaring it as unhealthy.</p>
</dd>
<dt><code translate="no">--var=<var translate="no">stringToString</var></code></dt>
<dd><p>Manifest variable substitution. Multiple can be set by using the flag multiple times (for example NAME=VALUE).</p>
</dd>
<dt><code translate="no">--vars-file=<var translate="no">stringArray</var></code></dt>
<dd><p>JSON or YAML file to read variable substitutions from. Can be supplied multiple times.</p>
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


