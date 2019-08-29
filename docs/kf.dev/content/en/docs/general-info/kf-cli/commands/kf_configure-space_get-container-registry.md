---
title: "kf configure-space get-container-registry"
slug: kf-configure-space-get-container-registry
url: /docs/general-info/kf-cli/commands/kf-configure-space-get-container-registry/
---
## kf configure-space get-container-registry

Get the container registry used for builds.

### Synopsis

Get the container registry used for builds.

```
kf configure-space get-container-registry SPACE_NAME [flags]
```

### Examples

```
  kf configure-space get-container-registry my-space
```

### Options

```
  -h, --help   help for get-container-registry
```

### Options inherited from parent commands

```
      --config string       Config file (default is $HOME/.kf)
      --kubeconfig string   Kubectl config file (default is $HOME/.kube/config)
      --log-http            Log HTTP requests to stderr
      --namespace string    Kubernetes namespace to target
```

### SEE ALSO

* [kf configure-space](/docs/general-info/kf-cli/commands/kf-configure-space/)	 - Set configuration for a space

