---
title: "Compare Cloud Foundry and Kf services"
weight: 20
---

This document provides a side-by-side comparison of the various services
available on Cloud Foundry (CF) and those that Kf integrates
with on Google Cloud.

{{< note >}} This guide does not attempt to compare the syntax or semantics of the SDK,
APIs, or command-line tools provided by CF and Kf.{{< /note >}}

<table>
  <thead>
    <tr>
      <th>Service category</th>
      <th>Service</th>
      <th>CF</th>
      <th>Kf</th>
    </tr>
  </thead>
  <tbody>
  <!-- Only include Service Category for the first row of each type -->

  <!-- Platform Components -->
  <tr>
    <td>Platform</td>
    <td>Infrastructure Orchestrator</td>
    <td>BOSH</td>
    <td>Kubernetes</td>
  </tr>
  <tr>
    <td></td>
    <td>PaaS</td>
    <td>CF Application Runtime (CFAR)</td>
    <td>Kf</td>
  </tr>

  <!-- Data Management -->
  <tr>
    <td>Data management</td>
    <td>Service Broker</td>
    <td>Service Broker Tile</td>
    <td>Kubernetes Deployed Service Brokers</td>
  </tr>
  <tr>
    <td></td>
    <td>MySQL</td>
    <td>MySQL Tile</td>
    <td>Kf Cloud Service Broker</td>
  </tr>
  <tr>
    <td></td>
    <td>MongoDB</td>
    <td>MongoDB Tile</td>
    <td>Kf Cloud Service Broker</td>
  </tr>
  <tr>
    <td></td>
    <td>RabbitMQ</td>
    <td>RabbitMQ Tile</td>
    <td>Kf Cloud Service Broker</td>
  </tr>
  <tr>
    <td></td>
    <td>Redis</td>
    <td>Redis Tile</td>
    <td>Kf Cloud Service Broker</td>
  </tr>
  <tr>
    <td></td>
    <td>Eureka</td>
    <td>Spring Cloud Services Tile</td>
    <td><a href="{{docs_path}}concepts/service-discovery">Service Discovery</a></td>
  </tr>
  <tr>
    <td></td>
    <td>Spring Cloud Config</td>
    <td>Spring Cloud Services Tile</td>
    <td><a href="{{docs_path}}how-to/deploying-spring-cloud-config">Spring Cloud Config</a></td>
  </tr>

  <!-- Operations tooling -->
  <tr>
    <td>Operations tooling</td>
    <td>Continuous Integration (CI)</td>
    <td>Concourse Tile</td>
    <td><a href="https://github.com/concourse/concourse-chart" target="_blank" class="external">Concourse Helm Chart</a></td>
  </tr>

  <!-- Logging -->
  <tr>
    <td>Logging</td>
    <td>Google Cloud</td>
    <td>Google Cloud Firehose Nozzle</td>
    <td><a href="https://cloud.google.com/blog/products/management-tools/using-logging-your-apps-running-kubernetes-engine" target="_blank" class="external">Google Cloud Logging Kubernetes Agent</a></td>
  </tr>
  <tr>
    <td></td>
    <td>Elastic</td>
    <td>Elastic Firehose Nozzle</td>
    <td><a href="https://cloud.google.com/solutions/partners/monitoring-gke-on-prem-with-the-elastic-stack">Elastic Stack Agent</a></td>
  </tr>
  <tr>
    <td></td>
    <td>Splunk</td>
    <td>Splunk Firehose Nozzle</td>
    <td><a href="https://cloud.google.com/solutions/logging-anthos-with-splunk-connect">Splunk Connect</a></td>
  </tr>
  <tr>
    <td></td>
    <td>Metrics</td>
    <td>CF App Metrics</td>
    <td><a href="https://cloud.google.com/monitoring/">Google Cloud Monitoring Kubernetes AGent</a></td>
  </tr>
  </tbody>
</table>
