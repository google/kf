---
title: "kf logs"
slug: kf-logs
url: /docs/general-info/kf-cli/commands/kf-logs/
---
## kf logs

View or follow logs for an app

### Synopsis

View or follow logs for an app

```
kf logs APP_NAME [flags]
```

### Examples

```
  kf logs myapp
  
  # Get the last 20 log lines
  kf logs myapp -n 20
  
  # Follow/tail the log stream
  kf logs myapp -f
```

### Options

```
  -f, --follow       Follow the log stream of the app.
  -h, --help         help for logs
  -n, --number int   Show the last N lines of logs. (default 10)
```

### Options inherited from parent commands

```
      --config string       Config file (default is $HOME/.kf)
      --kubeconfig string   Kubectl config file (default is $HOME/.kube/config)
      --namespace string    Kubernetes namespace to target
```

### SEE ALSO

* [kf](/docs/general-info/kf-cli/commands/kf/)	 - A MicroPaaS for Kubernetes with a Cloud Foundry style developer expeience

