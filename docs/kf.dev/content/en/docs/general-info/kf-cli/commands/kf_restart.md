---
title: "kf restart"
slug: kf-restart
url: /docs/general-info/kf-cli/commands/kf-restart/
---
## kf restart

Restarts all running instances of the app

### Synopsis

Restarts all running instances of the app

```
kf restart APP_NAME [flags]
```

### Examples

```
  kf restart myapp
```

### Options

```
      --async   Don't wait for the action to complete on the server before returning
  -h, --help    help for restart
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

