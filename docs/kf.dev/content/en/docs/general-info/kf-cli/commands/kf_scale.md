---
title: "kf scale"
slug: kf-scale
url: /docs/general-info/kf-cli/commands/kf-scale/
---
## kf scale

Change or view the instance count for an app

### Synopsis

Change or view the instance count for an app

```
kf scale APP_NAME [flags]
```

### Examples

```
  # Display current scale settings
  kf scale myapp
  # Scale to exactly 3 instances
  kf scale myapp --instances 3
  # Scale to at least 3 instances
  kf scale myapp --min 3
  # Scale between 0 and 5 instances
  kf scale myapp --max 5
  # Scale between 3 and 5 instances depending on traffic
  kf scale myapp --min 3 --max 5
```

### Options

```
  -h, --help            help for scale
  -i, --instances int   Number of instances. (default -1)
      --max int         Maximum number of instances to allow the autoscaler to scale to. 0 implies the app can be scaled to ∞. (default -1)
      --min int         Minimum number of instances to allow the autoscaler to scale to. 0 implies the app can be scaled to 0. (default -1)
```

### Options inherited from parent commands

```
      --config string       Config file (default is $HOME/.kf)
      --kubeconfig string   Kubectl config file (default is $HOME/.kube/config)
      --namespace string    Kubernetes namespace to target
```

### SEE ALSO

* [kf](/docs/general-info/kf-cli/commands/kf/)	 - A MicroPaaS for Kubernetes with a Cloud Foundry style developer expeience

