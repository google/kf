---
title: "kf-logs"
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
  kf logs myapp -n 20
  kf logs myapp -f
  
```

### Options

```
  -f, --follow       Follow the log stream of the app.
  -h, --help         help for logs
  -n, --number int   The number of lines from the end of the logs to show. (default 10)
```

### Options inherited from parent commands

```
      --config string       config file (default is $HOME/.kf)
      --kubeconfig string   kubectl config file (default is $HOME/.kube/config)
      --namespace string    kubernetes namespace
```

### SEE ALSO

* [kf](/docs/general-info/kf-cli/commands/kf/)	 - kf is like cf for Knative

