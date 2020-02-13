---
title: "kf doctor"
slug: kf-doctor
url: /docs/general-info/kf-cli/commands/kf-doctor/
---
## kf doctor

Doctor runs validation tests against one or more components

### Synopsis

Doctor runs tests one or more components to validate them.

 If no arguments are supplied, then all tests are run. If one or more arguments are suplied then only those components are run.

 Possible components are: buildpacks, cluster, istio

```
kf doctor [COMPONENT...] [flags]
```

### Examples

```
  kf doctor cluster
```

### Options

```
      --delay duration   Set the delay between executions (default 5s)
  -h, --help             help for doctor
      --retries int      Number of times to retry doctor if it isn't successful (default 1)
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

