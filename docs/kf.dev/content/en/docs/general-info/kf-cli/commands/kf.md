---
date: 2019-08-14T21:52:09-06:00
title: "kf"
slug: kf
url: /docs/general-info/kf-cli/commands/kf/
---
## kf

kf is like cf for Knative

### Synopsis

kf is like cf for Knative

```
kf [flags]
```

### Options

```
      --config string       config file (default is $HOME/.kf)
  -h, --help                help for kf
      --kubeconfig string   kubectl config file (default is $HOME/.kube/config)
      --namespace string    kubernetes namespace
```

### SEE ALSO

* [kf app](/docs/general-info/kf-cli/commands/kf_app/)	 - Get a pushed app
* [kf apps](/docs/general-info/kf-cli/commands/kf_apps/)	 - List pushed apps
* [kf bind-service](/docs/general-info/kf-cli/commands/kf_bind-service/)	 - Bind a service instance to an app
* [kf bindings](/docs/general-info/kf-cli/commands/kf_bindings/)	 - List bindings
* [kf build-logs](/docs/general-info/kf-cli/commands/kf_build-logs/)	 - Get the logs of the given build
* [kf buildpacks](/docs/general-info/kf-cli/commands/kf_buildpacks/)	 - List buildpacks in current builder
* [kf builds](/docs/general-info/kf-cli/commands/kf_builds/)	 - List the builds in the current space
* [kf completion](/docs/general-info/kf-cli/commands/kf_completion/)	 - 
* [kf configure-space](/docs/general-info/kf-cli/commands/kf_configure-space/)	 - Set configuration for a space
* [kf create-quota](/docs/general-info/kf-cli/commands/kf_create-quota/)	 - Create a quota
* [kf create-route](/docs/general-info/kf-cli/commands/kf_create-route/)	 - Create a route
* [kf create-service](/docs/general-info/kf-cli/commands/kf_create-service/)	 - Create a service instance
* [kf create-space](/docs/general-info/kf-cli/commands/kf_create-space/)	 - Create a space
* [kf debug](/docs/general-info/kf-cli/commands/kf_debug/)	 - Show debugging information useful for filing a bug report
* [kf delete](/docs/general-info/kf-cli/commands/kf_delete/)	 - Delete an existing app
* [kf delete-quota](/docs/general-info/kf-cli/commands/kf_delete-quota/)	 - Delete a quota
* [kf delete-route](/docs/general-info/kf-cli/commands/kf_delete-route/)	 - Delete a route
* [kf delete-service](/docs/general-info/kf-cli/commands/kf_delete-service/)	 - Delete a service instance
* [kf delete-space](/docs/general-info/kf-cli/commands/kf_delete-space/)	 - Delete a space
* [kf doctor](/docs/general-info/kf-cli/commands/kf_doctor/)	 - Doctor runs validation tests against one or more components
* [kf env](/docs/general-info/kf-cli/commands/kf_env/)	 - List the names and values of the environment variables for an app
* [kf install](/docs/general-info/kf-cli/commands/kf_install/)	 - Install kf
* [kf logs](/docs/general-info/kf-cli/commands/kf_logs/)	 - View or follow logs for an app
* [kf map-route](/docs/general-info/kf-cli/commands/kf_map-route/)	 - Map a route to an app
* [kf marketplace](/docs/general-info/kf-cli/commands/kf_marketplace/)	 - List available offerings in the marketplace
* [kf proxy](/docs/general-info/kf-cli/commands/kf_proxy/)	 - Creates a proxy to an app on a local port
* [kf push](/docs/general-info/kf-cli/commands/kf_push/)	 - Push a new app or sync changes to an existing app
* [kf quota](/docs/general-info/kf-cli/commands/kf_quota/)	 - Show quota info for a space
* [kf restage](/docs/general-info/kf-cli/commands/kf_restage/)	 - Restage creates a new container using the given source code and current buildpacks
* [kf restart](/docs/general-info/kf-cli/commands/kf_restart/)	 - Restart stops the current pods and create new ones
* [kf routes](/docs/general-info/kf-cli/commands/kf_routes/)	 - List routes in space
* [kf scale](/docs/general-info/kf-cli/commands/kf_scale/)	 - Change or view the instance count for an app
* [kf service](/docs/general-info/kf-cli/commands/kf_service/)	 - Show service instance info
* [kf services](/docs/general-info/kf-cli/commands/kf_services/)	 - List all service instances in the target namespace
* [kf set-env](/docs/general-info/kf-cli/commands/kf_set-env/)	 - Set an environment variable for an app
* [kf space](/docs/general-info/kf-cli/commands/kf_space/)	 - Show space info
* [kf spaces](/docs/general-info/kf-cli/commands/kf_spaces/)	 - List all kf spaces
* [kf stacks](/docs/general-info/kf-cli/commands/kf_stacks/)	 - List stacks in current builder
* [kf start](/docs/general-info/kf-cli/commands/kf_start/)	 - Start starts the app
* [kf stop](/docs/general-info/kf-cli/commands/kf_stop/)	 - Stop stops the app
* [kf target](/docs/general-info/kf-cli/commands/kf_target/)	 - Set or view the targeted space
* [kf unbind-service](/docs/general-info/kf-cli/commands/kf_unbind-service/)	 - Unbind a service instance from an app
* [kf unmap-route](/docs/general-info/kf-cli/commands/kf_unmap-route/)	 - Unmap a route from an app
* [kf unset-env](/docs/general-info/kf-cli/commands/kf_unset-env/)	 - Unset an environment variable for an app
* [kf update-quota](/docs/general-info/kf-cli/commands/kf_update-quota/)	 - Update a quota
* [kf vcap-services](/docs/general-info/kf-cli/commands/kf_vcap-services/)	 - Print the VCAP_SERVICES environment variable for an app
* [kf version](/docs/general-info/kf-cli/commands/kf_version/)	 - Display the CLI version

###### Auto generated by spf13/cobra on 14-Aug-2019
