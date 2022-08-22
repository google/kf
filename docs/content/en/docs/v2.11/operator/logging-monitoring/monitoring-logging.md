---
title: View logs
---

Kf provides you with several types of logs. This document describes these logs and how to access them.

## Application logs
All logs written to standard output `stdout` and standard error `stderr`, are uploaded to [Cloud Logging](https://cloud.google.com/logging) and stored under the log name `user-container`.

Open [Cloud Logging](https://cloud.google.com/logging/docs/view/logs-explorer-interface) and run the following query:
<pre class="devsite-click-to-copy" translate="no">
resource.type="k8s_container"
log_name="projects/<var>YOUR_PROJECT_ID</var>/logs/user-container"
resource.labels.project_id=<var>YOUR_PROJECT_ID</var>
resource.labels.location=<var>GCP_COMPUTE_ZONE (e.g. us-central1-a)</var>
resource.labels.cluster_name=<var>YOUR_CLUSTER_NAME</var>
resource.labels.namespace_name=<var>YOUR_KF_SPACE_NAME</var>
resource.labels.pod_name:<var>YOUR_KF_APP_NAME</var></pre>

You should see all your application logs written on standard `stdout` and standard error `stderr`.

## Access logs for your applications

Kf provides access logs using [Istio sidecar injection](https://cloud.google.com/service-mesh/docs/proxy-injection). Access logs are stored under the log name ``server-accesslog-stackdriver``.

Open [Cloud Logging](https://cloud.google.com/logging/docs/view/logs-explorer-interface) and run the following query:
<pre class="devsite-click-to-copy" translate="no">
resource.type="k8s_container"
log_name="projects/<var>YOUR_PROJECT_ID</var>/logs/server-accesslog-stackdriver"
resource.labels.project_id=<var>YOUR_PROJECT_ID</var>
resource.labels.location=<var>GCP_COMPUTE_ZONE (e.g. us-central1-a)</var>
resource.labels.cluster_name=<var>YOUR_CLUSTER_NAME</var>
resource.labels.namespace_name=<var>YOUR_KF_SPACE_NAME</var>
resource.labels.pod_name:<var>YOUR_KF_APP_NAME</var></pre>

You should see access logs for your application. Sample access log:

<pre class="devsite-click-to-copy" translate="no">
{
  "insertId": "166tsrsg273q5mf",
  "httpRequest": {
    "requestMethod": "GET",
    "requestUrl": "http://test-app-38n6dgwh9kx7h-c72edc13nkcm.***. ***.nip.io/",
    "requestSize": "738",
    "status": 200,
    "responseSize": "3353",
    "remoteIp": "10.128.0.54:0",
    "serverIp": "10.48.0.18:8080",
    "latency": "0.000723777s",
    "protocol": "http"
  },
  "resource": {
    "type": "k8s_container",
    "labels": {
      "container_name": "user-container",
      "project_id": ***,
      "namespace_name": ***,
      "pod_name": "test-app-85888b9796-bqg7b",
      "location": "us-central1-a",
      "cluster_name": ***
    }
  },
  "timestamp": "2020-11-19T20:09:21.721815Z",
  "severity": "INFO",
  "labels": {
    "source_canonical_service": "istio-ingressgateway",
    "source_principal": "spiffe://***.svc.id.goog/ns/istio-system/sa/istio-ingressgateway-service-account",
    "request_id": "0e3bac08-ab68-408f-9b14-0aec671845bf",
    "source_app": "istio-ingressgateway",
    "response_flag": "-",
    "route_name": "default",
    "upstream_cluster": "inbound|80|http-user-port|test-app.***.svc.cluster.local",
    "destination_name": "test-app-85888b9796-bqg7b",
    "destination_canonical_revision": "latest",
    "destination_principal": "spiffe://***.svc.id.goog/ns/***/sa/sa-test-app",
    "connection_id": "82261",
    "destination_workload": "test-app",
    "destination_namespace": ***,
    "destination_canonical_service": "test-app",
    "upstream_host": "127.0.0.1:8080",
    "log_sampled": "false",
    "mesh_uid": "proj-228179605852",
    "source_namespace": "istio-system",
    "requested_server_name": "outbound_.80_._.test-app.***.svc.cluster.local",
    "source_canonical_revision": "asm-173-6",
    "x-envoy-original-dst-host": "",
    "destination_service_host": "test-app.***.svc.cluster.local",
    "source_name": "istio-ingressgateway-5469f77856-4n2pw",
    "source_workload": "istio-ingressgateway",
    "x-envoy-original-path": "",
    "service_authentication_policy": "MUTUAL_TLS",
    "protocol": "http"
  },
  "logName": "projects/*/logs/server-accesslog-stackdriver",
  "receiveTimestamp": "2020-11-19T20:09:24.627065813Z"
}
</pre>

## Audit logs

[Audit Logs](https://cloud.google.com/kubernetes-engine/docs/how-to/audit-logging) provides a
chronological record of calls that have been made to the Kubernetes API Server. Kubernetes audit log entries are useful for investigating suspicious API requests, for collecting statistics, or for creating monitoring alerts for unwanted API calls.

Open [Cloud Logging](https://cloud.google.com/logging/docs/view/logs-explorer-interface) and run the following query:
<pre class="devsite-click-to-copy" translate="no">
resource.type="k8s_container"
log_name="projects/<var>YOUR_PROJECT_ID</var>/logs/cloudaudit.googleapis.com%2Factivity"
resource.labels.project_id=<var>YOUR_PROJECT_ID</var>
resource.labels.location=<var>GCP_COMPUTE_ZONE (e.g. us-central1-a)</var>
resource.labels.cluster_name=<var>YOUR_CLUSTER_NAME</var>
protoPayload.request.metadata.name=<var>YOUR_APP_NAME</var>
protoPayload.methodName:"deployments."</pre>

You should see a trace of calls being made to the Kubernetes API server.


## Configure logging access control

Follow [these instructions](https://cloud.google.com/logging/docs/access-control) to provide logs access to developers and other members on the team. The role `roles/logging.viewer` provides read-only access to logs.

## Use Logs Router

You can also use [Logs Router](https://cloud.google.com/logging/docs/routing/overview) to route the logs to supported destinations.
