---
title: "kf add-service-broker"
slug: kf-add-service-broker
url: /docs/general-info/kf-cli/commands/kf-add-service-broker/
---
## kf add-service-broker

Add a cluster service broker to service catalog

### Synopsis

Add a cluster service broker to service catalog

```
kf add-service-broker BROKER_NAME URL [flags]
```

### Examples

```
  kf add-service-broker mybroker http://mybroker.broker.svc.cluster.local
```

### Options

```
  -h, --help   help for add-service-broker
```

### Options inherited from parent commands

```
      --config string       Config file (default is $HOME/.kf)
      --kubeconfig string   Kubectl config file (default is $HOME/.kube/config)
      --namespace string    Kubernetes namespace to target
```

### SEE ALSO

* [kf](/docs/general-info/kf-cli/commands/kf/)	 - A MicroPaaS for Kubernetes with a Cloud Foundry style developer expeience

