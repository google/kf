---
title: "kf logs"
slug: kf-logs
url: /docs/general-info/kf-cli/commands/kf-logs/
---
## kf logs

Tail or show logs for an app

### Synopsis

Tail or show logs for an app

```
kf logs APP_NAME [flags]
```

### Examples

```
  # Follow/tail the log stream
  kf logs myapp
  
  # Follow/tail the log stream with 20 lines of context
  kf logs myapp -n 20
  
  # Get recent logs from the app
  kf logs myapp --recent
  
  # Get the most recent 200 lines of logs from the app
  kf logs myapp --recent -n 200
```

### Options

```
  -h, --help         help for logs
  -n, --number int   Show the last N lines of logs. (default 10)
      --recent       Dump recent logs instead of tailing.
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

