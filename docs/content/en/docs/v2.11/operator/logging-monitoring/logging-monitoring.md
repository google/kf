---
title: Logging and monitoring
---

Kf can use GKE's Google Cloud integrations
to send a log of events to your Cloud Monitoring and Cloud Logging project for observability. For more information, see
[Overview of GKE operations](https://cloud.google.com/monitoring/kubernetes-engine).

Kf deploys two server side components:

1. Controller
2. Webhook

To view the logs for these components, use the following Cloud Logging query:

```
resource.type="k8s_container"
resource.labels.project_id=<PROJECT ID>
resource.labels.location=<GCP ZONE>
resource.labels.cluster_name=<CLUSTER NAME>
resource.labels.namespace_name="kf"
labels.k8s-pod/app=<controller OR webhook>
```

{{< note >}} Replace the values inside the <> with the correct values.{{< /note >}}
