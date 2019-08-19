---
title: "kf bind-service"
slug: kf-bind-service
url: /docs/general-info/kf-cli/commands/kf-bind-service/
---
## kf bind-service

Bind a service instance to an app

### Synopsis

Bind a service instance to an app

```
kf bind-service APP_NAME SERVICE_INSTANCE [-c PARAMETERS_AS_JSON] [--binding-name BINDING_NAME] [flags]
```

### Examples

```
  kf bind-service myapp mydb -c '{"permissions":"read-only"}'
```

### Options

```
  -b, --binding-name string   name to expose service instance to app process with (default: service instance name)
  -h, --help                  help for bind-service
```

### Options inherited from parent commands

```
      --config string       config file (default is $HOME/.kf)
      --kubeconfig string   kubectl config file (default is $HOME/.kube/config)
      --namespace string    kubernetes namespace
```

### SEE ALSO

* [kf](/docs/general-info/kf-cli/commands/kf/)	 - kf is like cf for Knative

