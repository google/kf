---
title: "kf configure-space set-env"
slug: kf-configure-space-set-env
url: /docs/general-info/kf-cli/commands/kf-configure-space-set-env/
---
## kf configure-space set-env

Set a space-wide environment variable.

### Synopsis

Set a space-wide environment variable.

```
kf configure-space set-env SPACE_NAME ENV_VAR_NAME ENV_VAR_VALUE [flags]
```

### Examples

```
  kf configure-space set-env my-space ENVIRONMENT production
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

* [kf configure-space](/docs/general-info/kf-cli/commands/kf-configure-space/)	 - Set configuration for a space

