---
title: "kf create-service-broker"
slug: kf-create-service-broker
url: /docs/general-info/kf-cli/commands/kf-create-service-broker/
---
## kf create-service-broker

Add a cluster service broker to service catalog

### Synopsis

Add a cluster service broker to service catalog

```
kf create-service-broker BROKER_NAME URL [flags]
```

### Examples

```
  kf create-service-broker mybroker http://mybroker.broker.svc.cluster.local
```

### Options

```
  -h, --help           help for create-service-broker
      --space-scoped   Set to create a space scoped service broker.
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

