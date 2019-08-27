---
title: "kf delete-service-broker"
slug: kf-delete-service-broker
url: /docs/general-info/kf-cli/commands/kf-delete-service-broker/
---
## kf delete-service-broker

Remove a cluster service broker from service catalog

### Synopsis

Remove a cluster service broker from service catalog

```
kf delete-service-broker BROKER_NAME [flags]
```

### Examples

```
  kf delete-service-broker mybroker
```

### Options

```
  -h, --help   help for delete-service-broker
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

