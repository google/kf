---
title: "kf delete-route"
slug: kf-delete-route
url: /docs/general-info/kf-cli/commands/kf-delete-route/
---
## kf delete-route

Delete a route

### Synopsis

Delete a route

```
kf delete-route DOMAIN [--hostname HOSTNAME] [--path PATH] [flags]
```

### Examples

```

  kf delete-route example.com --hostname myapp # myapp.example.com
  kf delete-route example.com --hostname myapp --path /mypath # myapp.example.com/mypath
  
```

### Options

```
  -h, --help              help for delete-route
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

