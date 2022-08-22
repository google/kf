---
title: Create and user monitoring dashboards.
---

You can use [Google Cloud Monitoring dashboards](https://cloud.google.com/monitoring/dashboards) to create custom dashboards and charts. Kf comes with a default template which can be used to create dashboards to monitor the performance of your applications.

## Application performance dashboard

Run the following commands to deploy a dashboard in your monitoring workspace in [Cloud monitoring dashboards](https://cloud.google.com/monitoring/dashboards) to monitor performance of your apps. This dashboard has application performance metrics like requests/sec, round trip latency, HTTP error codes, and more.

<pre class="devsite-terminal" suppresswarning="true" translate="no">
<code class="devsite-terminal">git clone https://github.com/google/kf</code>
<code class="devsite-terminal">cd ./kf/dashboards</code>
<code class="devsite-terminal">./create-dashboard.py my-dashboard my-cluster my-space</code>
</pre>

## System resources and performance dashboard

You can view all the system resources and performance metrics such as list of nodes, pods, containers and much more using a built-in dashboard. Click the link below to access the system dashboard.

  <a href="https://console.cloud.google.com/monitoring/dashboards/resourceList/kubernetes" target="_blank" class="button button-primary" track-type="tasks" track-name="consoleLink" track-metadata-position="body" track-metadata-end-goal="connectToSerialConsole" >
    System dashboard</a></p></li>
</ol>

More details about this dashboard can be found [here](https://cloud.google.com/stackdriver/docs/solutions/gke/observing).

## Create SLO and alerts

You can create [SLOs](https://cloud.google.com/stackdriver/docs/solutions/slo-monitoring/ui/create-slo) and [Alerts](https://cloud.google.com/stackdriver/docs/solutions/slo-monitoring/ui/create-alert) on available metrics to monitor performance and availability of both system and applications. For example, you can use the metrics `istio.io/service/server/response_latencies` to setup an alert on the application roundtrip latency.

## Configure dashboard access control

Follow [these instructions](https://cloud.google.com/monitoring/access-control) to provide dashboard access to developers and other members on the team. The role `roles/monitoring.dashboardViewer` provides read-only access to dashboards.
