---
title: Reduce deployment risk with blue-green deployments
description: How to deploy highly available applications.
weight: 300
---

{{<note>}}
Kf will perform rolling deployments by default if you upgrade an app in place by starting
an extra instance of it, waiting for it to become healthy, and then replacing an existing
instance.
{{</note>}}

This page shows you how to deploy a new version of your application and migrate
traffic over from an old to a new version.

## Push the initial App

Use the Kf CLI to push the initial version of your App
with any routes:

```sh
$ kf push app-v1 --route my-app.my-space.example.com
```

## Push the updated App

Use the Kf CLI to push a new version of your App without
any routes:

```sh
$ kf push app-v2 --no-route
```

## Add routes to the updated App

Use the Kf CLI to bind all existing routes to the updated
App with a weight of 0 to ensure that they don't get any requests:

```sh
$ kf map-route app-v2 my-space.example.com --hostname my-app --weight 0
```

## Shift traffic

Start shifting traffic from the old App to the updated App by updating the
weights on the routes:

```sh
$ kf map-route app-v1 my-space.example.com --hostname my-app --weight 80
$ kf map-route app-v2 my-space.example.com --hostname my-app --weight 20
```

If the deployment is going well, you can shift more traffic by updating the
weights again:

```sh
$ kf map-route app-v1 my-space.example.com --hostname my-app --weight 50
$ kf map-route app-v2 my-space.example.com --hostname my-app --weight 50
```

{{< note >}} Weights split traffic proportionally to the sum of the weights of all
Apps mapped to the route. It's a common practice to treat weights as percentages.{{< /note >}}


## Complete traffic shifting

After you're satisfied that the new service hasn't introduced regressions,
complete the rollout by shifting all traffic to the new instance:

```sh
$ kf map-route app-v1 my-space.example.com --hostname my-app --weight 0
$ kf map-route app-v2 my-space.example.com --hostname my-app --weight 100
```

## Turn down the original App

After you're satisfied that quick rollbacks aren't needed, remove the original
route and stop the App:

```sh
$ kf unmap-route app-v1 myspace.example.com --hostname my-app
$ kf stop app-v1
```

Or delete the App and all associated route mappings:

```sh
$ kf delete app-v1
```

