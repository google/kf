---
title: "kf update-quota"
slug: kf-update-quota
url: /docs/general-info/kf-cli/commands/kf-update-quota/
---
## kf update-quota

Update a quota

### Synopsis

Update a quota

```
kf update-quota SPACE_NAME [flags]
```

### Options

```
  -c, --cpu string      The total available CPU across all builds and applications in a space (e.g. 400m). Default: unlimited (default "undefined")
  -h, --help            help for update-quota
  -m, --memory string   The total available memory across all builds and applications in a space (e.g. 10Gi, 500Mi). Default: unlimited (default "undefined")
  -r, --routes string   The total number of routes that can exist in a space. Default: unlimited (default "undefined")
```

### Options inherited from parent commands

```
      --config string       config file (default is $HOME/.kf)
      --kubeconfig string   kubectl config file (default is $HOME/.kube/config)
      --namespace string    kubernetes namespace
```

### SEE ALSO

* [kf](/docs/general-info/kf-cli/commands/kf/)	 - kf is like cf for Knative

