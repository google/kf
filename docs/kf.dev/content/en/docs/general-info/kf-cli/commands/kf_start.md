---
title: "kf start"
slug: kf-start
url: /docs/general-info/kf-cli/commands/kf-start/
---
## kf start

Start a staged application

### Synopsis

Start a staged application

```
kf start APP_NAME [flags]
```

### Examples

```
  kf start myapp
```

### Options

```
      --async   Don't wait for the action to complete on the server before returning
  -h, --help    help for start
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

