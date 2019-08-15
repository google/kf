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
  kf set-env myapp FOO bar
```

### Options

```
  -h, --help   help for set-env
```

### Options inherited from parent commands

```
      --config string       config file (default is $HOME/.kf)
      --kubeconfig string   kubectl config file (default is $HOME/.kube/config)
      --namespace string    kubernetes namespace
```

### SEE ALSO

* [kf](/docs/general-info/kf-cli/commands/kf/)	 - kf is like cf for Knative

