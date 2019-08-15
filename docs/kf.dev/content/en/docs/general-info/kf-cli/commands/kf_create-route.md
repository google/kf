---
title: "kf-create-route"
slug: kf-create-route
url: /docs/general-info/kf-cli/commands/kf-create-route/
---
## kf create-route

Create a route

### Synopsis

Create a route

```
kf create-route DOMAIN [--hostname HOSTNAME] [--path PATH] [flags]
```

### Examples

```

  # Using namespace (instead of SPACE)
  kf create-route example.com --hostname myapp # myapp.example.com
  kf create-route --namespace myspace example.com --hostname myapp # myapp.example.com
  kf create-route example.com --hostname myapp --path /mypath # myapp.example.com/mypath

  # [DEPRECATED] Using SPACE to match 'cf'
  kf create-route myspace example.com --hostname myapp # myapp.example.com
  kf create-route myspace example.com --hostname myapp --path /mypath # myapp.example.com/mypath
  
```

### Options

```
  -h, --help              help for create-route
      --hostname string   Hostname for the route
      --path string       URL Path for the route
```

### Options inherited from parent commands

```
      --config string       config file (default is $HOME/.kf)
      --kubeconfig string   kubectl config file (default is $HOME/.kube/config)
      --namespace string    kubernetes namespace
```

### SEE ALSO

* [kf](/docs/general-info/kf-cli/commands/kf/)	 - kf is like cf for Knative

