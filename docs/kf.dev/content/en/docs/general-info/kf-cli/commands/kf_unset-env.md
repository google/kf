---
title: "kf-unset-env"
slug: kf-unset-env
url: /docs/general-info/kf-cli/commands/kf-unset-env/
---
## kf unset-env

Unset an environment variable for an app

### Synopsis

Unset an environment variable for an app

```
kf unset-env APP_NAME ENV_VAR_NAME [flags]
```

### Examples

```
  kf unset-env myapp FOO
```

### Options

```
  -h, --help   help for unset-env
```

### Options inherited from parent commands

```
      --config string       config file (default is $HOME/.kf)
      --kubeconfig string   kubectl config file (default is $HOME/.kube/config)
      --namespace string    kubernetes namespace
```

### SEE ALSO

* [kf](/docs/general-info/kf-cli/commands/kf/)	 - kf is like cf for Knative

