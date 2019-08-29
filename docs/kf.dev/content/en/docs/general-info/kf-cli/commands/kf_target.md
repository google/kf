---
title: "kf target"
slug: kf-target
url: /docs/general-info/kf-cli/commands/kf-target/
---
## kf target

Set or view the targeted space

### Synopsis

Set or view the targeted space

```
kf target [flags]
```

### Examples

```
  # See the current space
  kf target
  # Target a space
  kf target -s my-space
```

### Options

```
  -h, --help           help for target
  -s, --space string   Target the given space.
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

