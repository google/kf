---
title: "kf create-service"
slug: kf-create-service
url: /docs/general-info/kf-cli/commands/kf-create-service/
---
## kf create-service

Create a service instance

### Synopsis

Create a service instance

```
kf create-service SERVICE PLAN SERVICE_INSTANCE [-c PARAMETERS_AS_JSON] [flags]
```

### Examples

```
  kf create-service db-service silver mydb -c '{"ram_gb":4}'
  kf create-service db-service silver mydb -c ~/workspace/tmp/instance_config.json
```

### Options

```
  -h, --help   help for create-service
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

