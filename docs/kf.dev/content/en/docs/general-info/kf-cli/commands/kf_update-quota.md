---
title: "kf update-quota"
slug: kf-update-quota
url: /docs/general-info/kf-cli/commands/kf-update-quota/
---
## kf update-quota

Update the quota for a space

### Synopsis

Update the quota for a space

```
kf update-quota SPACE_NAME [-m MEMORY] [-r ROUTES] [-c CPU] [flags]
```

### Examples

```
  kf update-quota my-space --memory 100Gi --routes 50
```

### Options

```
  -c, --cpu string      Total amount of CPU the space can have (e.g. 400m) (default: unlimited) (default "undefined")
  -h, --help            help for update-quota
  -m, --memory string   Total amount of memory the space can have (e.g. 10Gi, 500Mi) (default: unlimited) (default "undefined")
  -r, --routes string   Maximum number of routes the space can have (default: unlimited) (default "undefined")
```

### Options inherited from parent commands

```
      --config string       Config file (default is $HOME/.kf)
      --kubeconfig string   Kubectl config file (default is $HOME/.kube/config)
      --namespace string    Kubernetes namespace to target
```

### SEE ALSO

* [kf](/docs/general-info/kf-cli/commands/kf/)	 - A MicroPaaS for Kubernetes with a Cloud Foundry style developer expeience

