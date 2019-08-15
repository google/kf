---
title: "kf-push"
slug: kf-push
url: /docs/general-info/kf-cli/commands/kf-push/
---
## kf push

Push a new app or sync changes to an existing app

### Synopsis

Push a new app or sync changes to an existing app

```
kf push APP_NAME [flags]
```

### Examples

```

  kf push myapp
  kf push myapp --container-registry gcr.io/myproject
  kf push myapp --buildpack my.special.buildpack # Discover via kf buildpacks
  kf push myapp --env FOO=bar --env BAZ=foo
  
```

### Options

```
  -b, --buildpack string            Skip the 'detect' buildpack step and use the given name.
      --container-registry string   The container registry to push containers. Required if not targeting a Kf space.
      --docker-image string         The docker image to deploy.
  -e, --env stringArray             Set environment variables. Multiple can be set by using the flag multiple times (e.g., NAME=VALUE).
      --grpc                        Setup the container to allow application to use gRPC.
  -u, --health-check-type string    Application health check type (http or port, default: port)
  -h, --help                        help for push
  -i, --instances int               the number of instances (default is 1) (default -1)
  -f, --manifest string             Path to manifest
      --max-scale int               the maximum number of instances the autoscaler will scale to (default -1)
      --min-scale int               the minium number of instances the autoscaler will scale to (default -1)
      --no-manifest                 Ignore the manifest file.
      --no-route                    Do not map a route to this app and remove routes from previous pushes of this app
      --no-start                    Do not start an app after pushing
  -p, --path string                 The path the source code lives. Defaults to current directory. (default ".")
      --random-route                Create a random route for this app if the app doesn't have a route.
      --route stringArray           Use the routes flag to provide multiple HTTP and TCP routes. Each route for this app is created if it does not already exist.
      --service-account string      The service account to enable access to the container registry
  -t, --timeout int                 Time (in seconds) allowed to elapse between starting up an app and the first healthy response from the app.
```

### Options inherited from parent commands

```
      --config string       config file (default is $HOME/.kf)
      --kubeconfig string   kubectl config file (default is $HOME/.kube/config)
      --namespace string    kubernetes namespace
```

### SEE ALSO

* [kf](/docs/general-info/kf-cli/commands/kf/)	 - kf is like cf for Knative

