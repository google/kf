---
title: "kf unmap-route"
slug: kf-unmap-route
url: /docs/general-info/kf-cli/commands/kf-unmap-route/
---
## kf unmap-route

Unmap a route from an app

### Synopsis

Unmap a route from an app

```
kf unmap-route APP_NAME DOMAIN [--hostname HOSTNAME] [--path PATH] [flags]
```

### Examples

```
  kf unmap-route myapp example.com --hostname myapp # myapp.example.com
  kf unmap-route --namespace myspace myapp example.com --hostname myapp # myapp.example.com
  kf unmap-route myapp example.com --hostname myapp --path /mypath # myapp.example.com/mypath
```

### Options

```
      --async             Don't wait for the action to complete on the server before returning
  -h, --help              help for unmap-route
      --hostname string   Hostname for the route
      --path string       URL Path for the route
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

