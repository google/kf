---
title: Logging and monitoring overview 
---

By default, Kf includes native integration with [Cloud Monitoring](https://cloud.google.com/monitoring) and [Cloud Logging](https://cloud.google.com/logging/docs). When you create a cluster, both Monitoring and Cloud Logging are enabled by default. This integration lets you monitor your running clusters and help analyze your system and application performance using advanced profiling and tracing
capabilities.

Application level performance metrics is provided by [Istio sidecar injection](https://cloud.google.com/service-mesh/docs/proxy-injection) which is injected alongside applications deployed via Kf. You can also create [SLO](https://cloud.google.com/stackdriver/docs/solutions/slo-monitoring/ui/create-slo) and [Alerts](https://cloud.google.com/stackdriver/docs/solutions/slo-monitoring/ui/create-alert) using this default integration to monitor performance and availability of both system and applications.

Ensure the following are setup on your cluster:

- [Cloud Monitoring](https://cloud.google.com/monitoring) and [Cloud Logging](https://cloud.google.com/logging/docs) are enabled on the Kf cluster by default unless you disabled them explicitly, so no extra step is required.

- [Istio sidecar injection](https://cloud.google.com/service-mesh/docs/proxy-injection) is enabled.  Sidecar proxy will inject application level [performance metrics](https://cloud.google.com/monitoring/api/metrics_istio).
