---
title: "kf push"
slug: kf-push
url: /docs/general-info/kf-cli/commands/kf-push/
---
## kf push

Create a new app or sync changes to an existing app

### Synopsis

Create a new app or sync changes to an existing app

```
kf push APP_NAME [flags]
```

### Examples

```
  kf push myapp
  kf push myapp --buildpack my.special.buildpack # Discover via kf buildpacks
  kf push myapp --env FOO=bar --env BAZ=foo
  kf push myapp --stack cloudfoundry/cflinuxfs3 # Use a cflinuxfs3 runtime
```

### Options

```
      --args stringArray            Overwrite the args for the image. Can't be used with the command flag.
  -b, --buildpack string            Skip the 'detect' buildpack step and use the given name.
  -c, --command string              Startup command for the app, this overrides the default command specified by the web process.
      --container-registry string   Container registry to push sources to. Required for buildpack builds not targeting a Kf space.
      --docker-image string         Docker image to deploy.
      --enable-http2                Setup the container to allow application to use HTTP2 and gRPC.
      --entrypoint string           Overwrite the default entrypoint of the image. Can't be used with the command flag.
  -e, --env stringArray             Set environment variables. Multiple can be set by using the flag multiple times (e.g., NAME=VALUE).
  -u, --health-check-type string    Application health check type (http or port, default: port)
  -h, --help                        help for push
  -i, --instances int               Number of instances of the app to run (default: 1) (default -1)
  -f, --manifest string             Path to manifest
      --max-scale int               Maximum number of instances the autoscaler will scale to (default -1)
      --min-scale int               Minium number of instances the autoscaler will scale to (default -1)
      --no-manifest                 Ignore the manifest file.
      --no-route                    Do not map a route to this app and remove routes from previous pushes of this app
      --no-start                    Do not start an app after pushing
  -p, --path string                 Path to the source code (default: current directory) (default ".")
      --random-route                Create a random route for this app if the app doesn't have a route.
      --route stringArray           Use the routes flag to provide multiple HTTP and TCP routes. Each route for this app is created if it does not already exist.
  -s, --stack string                Base image to use for to use for apps created with a buildpack.
  -t, --timeout int                 Time (in seconds) allowed to elapse between starting up an app and the first healthy response from the app.
```

### Options inherited from parent commands

```
      --config string       Config file (default is $HOME/.kf)
      --kubeconfig string   Kubectl config file (default is $HOME/.kube/config)
      --log-http            Log HTTP requests to stderr
      --namespace string    Kubernetes namespace to target
```

### SEE ALSO

* [kf](/docs/general-info/kf-cli/commands/kf/)	 - A MicroPaaS for Kubernetes with a Cloud Foundry style developer expeience

