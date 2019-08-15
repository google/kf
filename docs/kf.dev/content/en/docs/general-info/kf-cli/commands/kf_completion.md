---
title: "kf completion"
slug: kf-completion
url: /docs/general-info/kf-cli/commands/kf-completion/
---
## kf completion



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
      --config string       config file (default is $HOME/.kf)
      --kubeconfig string   kubectl config file (default is $HOME/.kube/config)
      --namespace string    kubernetes namespace
```

### SEE ALSO

* [kf](/docs/general-info/kf-cli/commands/kf/)	 - kf is like cf for Knative

