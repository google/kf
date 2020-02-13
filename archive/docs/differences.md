# Differences between `kf` and `cf`

This document notes deviations in behavior between `kf` and `cf` for
user-facing tasks. For example, using Istio rather than the gorouter for traffic
wouldn't be mentioned, but any visible side-effects that causes would be.

The differences are broken down by command/workflow.

## Push

* If the CLI disconnects during a build in `kf` the app may not be updated
  whereas in `cf` it might.
