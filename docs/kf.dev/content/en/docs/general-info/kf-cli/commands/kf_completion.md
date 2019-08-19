---
title: "kf completion"
slug: kf-completion
url: /docs/general-info/kf-cli/commands/kf-completion/
---
## kf completion

Generate auto-completion files for kf commands

### Synopsis

completion is used to create set up bash/zsh auto-completion for kf commands.

```
kf completion bash|zsh [flags]
```

### Examples

```
  eval "$(kf completion bash)"
  eval "$(kf completion zsh)"
```

### Options

```
  -h, --help   help for completion
```

### Options inherited from parent commands

```
      --config string       Config file (default is $HOME/.kf)
      --kubeconfig string   Kubectl config file (default is $HOME/.kube/config)
      --namespace string    Kubernetes namespace to target
```

### SEE ALSO

* [kf](/docs/general-info/kf-cli/commands/kf/)	 - A MicroPaaS for Kubernetes with a Cloud Foundry style developer expeience

