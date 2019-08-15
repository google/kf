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

  kf scale myapp # Displays current scaling
  kf scale myapp -i 3 # Scale to exactly 3 instances
  kf scale myapp --instances 3 # Scale to exactly 3 instances
  kf scale myapp --min 3 # Autoscaler won't scale below 3 instances
  kf scale myapp --max 5 # Autoscaler won't scale above 5 instances
  kf scale myapp --min 3 --max 5 # Autoscaler won't below 3 or above 5 instances
  
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
      --config string       config file (default is $HOME/.kf)
      --kubeconfig string   kubectl config file (default is $HOME/.kube/config)
      --namespace string    kubernetes namespace
```

### SEE ALSO

* [kf](/docs/general-info/kf-cli/commands/kf/)	 - kf is like cf for Knative

