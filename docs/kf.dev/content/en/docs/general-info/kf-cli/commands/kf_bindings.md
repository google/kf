---
title: "kf bindings"
slug: kf-bindings
url: /docs/general-info/kf-cli/commands/kf-bindings/
---
## kf bindings

List bindings

### Synopsis

List bindings

```
kf bindings [--app APP_NAME] [--service SERVICE_NAME] [flags]
```

### Examples

```
  # Show all bindings
  kf bindings
  
  # Show bindings for "my-app"
  kf bindings --app my-app
  
  # Show bindings for a particular service
  kf bindings --service users-db
```

### Options

```
  -a, --app string       App to display bindings for
  -h, --help             help for bindings
  -s, --service string   Service instance to display bindings for
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

