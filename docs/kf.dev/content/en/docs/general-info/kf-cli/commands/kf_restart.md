---
title: "kf restart"
slug: kf-restart
url: /docs/general-info/kf-cli/commands/kf-restart/
---
## kf restart

Restart stops the current pods and create new ones

### Synopsis

Restart stops the current pods and create new ones

```
kf restart APP_NAME [flags]
```

### Examples

```

  kf restart myapp
  
```

### Options

```
  -h, --help   help for restart
```

### Options inherited from parent commands

```
      --config string       config file (default is $HOME/.kf)
      --kubeconfig string   kubectl config file (default is $HOME/.kube/config)
      --namespace string    kubernetes namespace
```

### SEE ALSO

* [kf](/docs/general-info/kf-cli/commands/kf/)	 - kf is like cf for Knative

