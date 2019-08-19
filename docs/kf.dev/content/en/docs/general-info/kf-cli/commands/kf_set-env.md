---
title: "kf set-env"
slug: kf-set-env
url: /docs/general-info/kf-cli/commands/kf-set-env/
---
## kf set-env

Set an environment variable for an app

### Synopsis

Set an environment variable for an app

```
kf set-env APP_NAME ENV_VAR_NAME ENV_VAR_VALUE [flags]
```

### Examples

```
  kf set-env myapp ENV production
```

### Options

```
  -h, --help   help for set-env
```

### Options inherited from parent commands

```
      --config string       Config file (default is $HOME/.kf)
      --kubeconfig string   Kubectl config file (default is $HOME/.kube/config)
      --namespace string    Kubernetes namespace to target
```

### SEE ALSO

* [kf](/docs/general-info/kf-cli/commands/kf/)	 - A MicroPaaS for Kubernetes with a Cloud Foundry style developer expeience

