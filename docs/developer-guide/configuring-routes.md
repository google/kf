# Configuring Routes

Some applications are useful without being accessible on the Internet, but most probably need to be available outside of the cluster at one or more HTTP endpoints. In Kf, this is the job of routes.

By default, each application is accessible to other applications in the cluster at the internal URI: `app-name.space-name.cluser.internal`. You can use these URIs when you deploy one or more applications in a cluster that need to directly communicate with one another; they allow traffic to go directly from one app to another rather than out of the cluster and back. This makes communications more secure, faster, and guaranteed to use the service in the local cluster.

If your app needs to be available outside of the cluster, you'll need to create routes for them.

## The Internal URI

The internal route for each application has some special characteristics.

* It always points to the most recent ready version of your app.
* Using it in your apps allows East West (point to point) routing.
* Traffic sent to it is load-balanced between running instances of your app.
* Traffic sent through it is used to determine if your app needs to scale up or down.

Routes allow you to create vanity URLs on top of the internal URI.

## Route Capabilities

Routes tell the cluster's ingress gateway where to deliver traffic and what to do if no apps are available on the given address.
By default, if no app is available on a route and the route receives a request it returns an [HTTP 503 status code](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/503).

Routes are comprised of three parts: host, domain, and path. For example in the URI `payroll.mydatacenter.example.com/login` 

* The host is `payroll` 
* The domain is `mydatacenter.example.com` 
* The path is `/login`

Routes must contain a host and domain, but the path is optional. Multiple routes can share the same host and domain if they specify different paths.
Multiple apps can share the same route and traffic will be routed to one of them. This is useful if you need to support legacy blue/green deployments.
If multiple apps are bound to different paths, the priority is longest to shortest path.

Some things routes don't currently allow:

* TCP traffic routing (pure L3 routing)
* Custom status codes
* Fault injection

## Using Routes

The following sections describe how to use the `kf` CLI to manage routes.

### List Routes

Developers can list routes for the current space using the `kf routes` command.

```.sh
$ kf routes
Getting routes in namespace: my-space
Found 2 routes in namespace my-space

HOST    DOMAIN       PATH    APPS
echo    example.com  /       echo
*       example.com  /login  uaa
```

### Create Route

Developers can create routes using the `kf create-route` command.

```.sh
# Create a route in the targeted space to match traffic for myapp.example.com/*
$ kf create-route example.com --hostname myapp

# Create a route in the space myspace to match traffic for myapp.example.com/*
$ kf create-route -n myspace example.com --hostname myapp

# Create a route in the targeted space to match traffic for myapp.example.com/mypath*
$ kf create-route example.com --hostname myapp --path /mypath

# [DEPRECATED] You can also supply the space name as the first parameter if you have
# scripts that rely on the old cf style API.
$ kf create-route myspace example.com --hostname myapp # myapp.example.com
```

AFter a route is created, if no apps are bound to it then a an [HTTP 503 status code](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/503)
is returned for any matching requests.

NOTE: Routes that share the same host and domain must be in the same space.

### Check Routes

Kf does not yet support checking routes. There is an [open issue](https://github.com/google/kf/issues/336) with more information.

### Map a Route to Your App

Developers can make their app accessible on a route using the `kf map-route` command.

```.sh
$ kf map-route MYAPP mycluster.example.com --host myapp --path mypath
```

### Unmap a Route

Developers can remove their app from being accessible on a route using the `kf unmap-route` command.

```.sh
$ kf unmap-route MYAPP mycluster.example.com --host myapp --path mypath
```

### Delete a Route

Developers can delete a route using the `kf delete-route` command.

```.sh
$ kf delete-route mycluster.example.com --host myapp --path mypath
```

Deleting a route will stop traffic from being routed to all applications listening on the route.

NOTE: If no other routes exist for the given host domain pair then another space can start to use the route.

### Declarative Routes in Your App Manifest

Routes can be managed declaratively in your app manifest file. They will be created if they do not yet exist.

```.yaml
---
applications:
- name: my-app
  # ...
  routes:
  - route: example.com
  - route: www.example.com/path
  - route: tcp.example.com:1234
```

NOTE: declaring routes in your manifest file will only create new routes, it will not delete routes you
created manually or as part of a previous push.
