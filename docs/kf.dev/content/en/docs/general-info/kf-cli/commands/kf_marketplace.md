---
title: "kf marketplace"
slug: kf-marketplace
url: /docs/general-info/kf-cli/commands/kf-marketplace/
---
## kf marketplace

List available offerings in the marketplace

### Synopsis

List available offerings in the marketplace

```
kf marketplace [-s SERVICE] [flags]
```

### Examples

```
  # Show services available in the marketplace
  kf marketplace
  
  # Show the plans available to a particular service
  kf marketplace -s google-storage
```

### Options

```
  -h, --help             help for marketplace
  -s, --service string   Show plan details for a particular service offering.
```

### Options inherited from parent commands

```
      --config string       Config file (default is $HOME/.kf)
      --kubeconfig string   Kubectl config file (default is $HOME/.kube/config)
      --namespace string    Kubernetes namespace to target
```

### SEE ALSO

* [kf](/docs/general-info/kf-cli/commands/kf/)	 - A MicroPaaS for Kubernetes with a Cloud Foundry style developer expeience

