---
title: "kf build-logs"
slug: kf-build-logs
url: /docs/general-info/kf-cli/commands/kf-build-logs/
---
## kf build-logs

Get the logs of the given build

### Synopsis

Get the logs of the given build

```
kf build-logs BUILD_NAME [flags]
```

### Examples

```
  kf build-logs build-12345
```

### Options

```
  -h, --help   help for build-logs
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

