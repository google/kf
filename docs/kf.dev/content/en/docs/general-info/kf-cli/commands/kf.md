---
title: "kf"
slug: kf
url: /docs/general-info/kf-cli/commands/kf/
---
## kf

A MicroPaaS for Kubernetes with a Cloud Foundry style developer expeience

### Synopsis

Kf is a MicroPaaS for Kubernetes with a Cloud Foundry style developer expeience.

 Kf aims to be fully compatible with Cloud Foundry applications and lifecycle. It supports logs, buildpacks, app manifests, routing, service brokers, and injected services.

 At the same time, it aims to improve the operational experience by supporting git-ops, self-healing infrastructure, containers, a service mesh, autoscaling, scale-to-zero, improved quota management and does it all on Kubernetes using industry-standard OSS tools including Knative, Istio, and Tekton.

```
kf [flags]
```

### Options

```
      --config string       Config file (default is $HOME/.kf)
  -h, --help                help for kf
      --kubeconfig string   Kubectl config file (default is $HOME/.kube/config)
      --log-http            Log HTTP requests to stderr
      --namespace string    Kubernetes namespace to target
```

### SEE ALSO

* [kf app](/docs/general-info/kf-cli/commands/kf-app/)	 - Print information about a deployed app
* [kf apps](/docs/general-info/kf-cli/commands/kf-apps/)	 - List pushed apps
* [kf bind-service](/docs/general-info/kf-cli/commands/kf-bind-service/)	 - Bind a service instance to an app
* [kf bindings](/docs/general-info/kf-cli/commands/kf-bindings/)	 - List bindings
* [kf build-logs](/docs/general-info/kf-cli/commands/kf-build-logs/)	 - Get the logs of the given build
* [kf buildpacks](/docs/general-info/kf-cli/commands/kf-buildpacks/)	 - List buildpacks in current builder
* [kf builds](/docs/general-info/kf-cli/commands/kf-builds/)	 - List the builds in the current space
* [kf completion](/docs/general-info/kf-cli/commands/kf-completion/)	 - Generate auto-completion files for kf commands
* [kf configure-space](/docs/general-info/kf-cli/commands/kf-configure-space/)	 - Set configuration for a space
* [kf create-route](/docs/general-info/kf-cli/commands/kf-create-route/)	 - Create a route
* [kf create-service](/docs/general-info/kf-cli/commands/kf-create-service/)	 - Create a service instance
* [kf create-service-broker](/docs/general-info/kf-cli/commands/kf-create-service-broker/)	 - Add a service broker to service catalog
* [kf create-space](/docs/general-info/kf-cli/commands/kf-create-space/)	 - Create a space
* [kf debug](/docs/general-info/kf-cli/commands/kf-debug/)	 - Show debugging information useful for filing a bug report
* [kf delete](/docs/general-info/kf-cli/commands/kf-delete/)	 - Delete an existing app
* [kf delete-quota](/docs/general-info/kf-cli/commands/kf-delete-quota/)	 - Remove all quotas for the space
* [kf delete-route](/docs/general-info/kf-cli/commands/kf-delete-route/)	 - Delete a route
* [kf delete-service](/docs/general-info/kf-cli/commands/kf-delete-service/)	 - Delete a service instance
* [kf delete-service-broker](/docs/general-info/kf-cli/commands/kf-delete-service-broker/)	 - Remove a service broker from service catalog
* [kf delete-space](/docs/general-info/kf-cli/commands/kf-delete-space/)	 - Delete a space
* [kf doctor](/docs/general-info/kf-cli/commands/kf-doctor/)	 - Doctor runs validation tests against one or more components
* [kf env](/docs/general-info/kf-cli/commands/kf-env/)	 - List the names and values of the environment variables for an app
* [kf install](/docs/general-info/kf-cli/commands/kf-install/)	 - Install kf
* [kf logs](/docs/general-info/kf-cli/commands/kf-logs/)	 - Tail or show logs for an app
* [kf map-route](/docs/general-info/kf-cli/commands/kf-map-route/)	 - Map a route to an app
* [kf marketplace](/docs/general-info/kf-cli/commands/kf-marketplace/)	 - List available offerings in the marketplace
* [kf proxy](/docs/general-info/kf-cli/commands/kf-proxy/)	 - Create a proxy to an app on a local port
* [kf push](/docs/general-info/kf-cli/commands/kf-push/)	 - Create a new app or sync changes to an existing app
* [kf quota](/docs/general-info/kf-cli/commands/kf-quota/)	 - Show quota info for a space
* [kf restage](/docs/general-info/kf-cli/commands/kf-restage/)	 - Rebuild and deploy using the last uploaded source code and current buildpacks
* [kf restart](/docs/general-info/kf-cli/commands/kf-restart/)	 - Restarts all running instances of the app
* [kf routes](/docs/general-info/kf-cli/commands/kf-routes/)	 - List routes in space
* [kf scale](/docs/general-info/kf-cli/commands/kf-scale/)	 - Change or view the instance count for an app
* [kf service](/docs/general-info/kf-cli/commands/kf-service/)	 - Show service instance info
* [kf services](/docs/general-info/kf-cli/commands/kf-services/)	 - List service instances
* [kf set-env](/docs/general-info/kf-cli/commands/kf-set-env/)	 - Set an environment variable for an app
* [kf space](/docs/general-info/kf-cli/commands/kf-space/)	 - Show space info
* [kf spaces](/docs/general-info/kf-cli/commands/kf-spaces/)	 - List all kf spaces
* [kf stacks](/docs/general-info/kf-cli/commands/kf-stacks/)	 - List stacks available in the space
* [kf start](/docs/general-info/kf-cli/commands/kf-start/)	 - Start a staged application
* [kf stop](/docs/general-info/kf-cli/commands/kf-stop/)	 - Stop a running application
* [kf target](/docs/general-info/kf-cli/commands/kf-target/)	 - Set or view the targeted space
* [kf unbind-service](/docs/general-info/kf-cli/commands/kf-unbind-service/)	 - Unbind a service instance from an app
* [kf unmap-route](/docs/general-info/kf-cli/commands/kf-unmap-route/)	 - Unmap a route from an app
* [kf unset-env](/docs/general-info/kf-cli/commands/kf-unset-env/)	 - Unset an environment variable for an app
* [kf update-quota](/docs/general-info/kf-cli/commands/kf-update-quota/)	 - Update the quota for a space
* [kf vcap-services](/docs/general-info/kf-cli/commands/kf-vcap-services/)	 - Print the VCAP_SERVICES environment variable for an app
* [kf version](/docs/general-info/kf-cli/commands/kf-version/)	 - Display the CLI version

