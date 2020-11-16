# Kf Dashboards

### Script

The [create-dashboard.py](create-dashboard.py) script is a Python script that
will use the [dashboard-template.json](dashboard-template.json) file to create
a template. It uses the currently targeted Google Cloud project (i.e., `gcloud config
get-value project`).

It takes the following 3 arguments:

1. Dashboard name - The name of the dashboard.
1. Cluster name - The name of the Kubernetes cluster.
1. Space - The name of the Kf space.

Note: Running the script is not idempotent.

Example:

```
./create-dashboard.py my-dashboard my-cluster my-space
```

### Templates

The JSON templates are created using steps found in the [Cloud Monitoring
Dashboard Samples](https://github.com/GoogleCloudPlatform/monitoring-dashboard-samples/blob/master/README.md).

They each contain the following sentinel values:
* `XXX-DASHBOARD-XXX` - The name of the dashboard.
* `XXX-CLUSTER-XXX` - The name of the Kubernetes cluster.
* `XXX-SPACE-XXX` - The name of the Kf Space.

If you plan to use this template directly (instead of using the script), be
sure to replace each instance of each sentinel value.
