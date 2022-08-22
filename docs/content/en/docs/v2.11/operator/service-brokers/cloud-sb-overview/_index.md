---
title: "About Kf Cloud Service Broker"
---

The Kf Cloud Service Broker is a Service Broker bundle that includes the open source
[Cloud Service Broker](https://github.com/cloudfoundry-incubator/cloud-service-broker)
and [Google Cloud Brokerpak](https://github.com/cloudfoundry-incubator/csb-brokerpak-gcp).
It is made available as a public Docker image and ready to deploy as a
Kubernetes service in Kf clusters. Once the
Kf Cloud Service Broker service is deployed in a cluster, developers can provision Google Cloud
backing services through the Kf Cloud Service Broker service, and bind the backing services to Kf Apps.

{{< note >}} Kf Cloud Service Brokeris not customizable, and the default
Google Cloud Brokerpak is included. If you would like to use a custom
Brokerpak, you can follow the steps in the
[open source Cloud Service Broker Google Cloud installation guide](https://github.com/cloudfoundry-incubator/csb-brokerpak-gcp/blob/main/docs/gcp-installation.md).
{{< /note >}}

## Requirements

* Kf Cloud Service Broker requires a MySQL instance and a service account for accessing the MySQL instance and Google Cloud backing services to be provisioned. Connection from the Kf Cloud Service Broker to the MySQL instance goes through the [Cloud SQL Auth Proxy](https://cloud.google.com/sql/docs/mysql/sql-proxy).
* Requests to access Google Cloud services (for example: Cloud SQL for MySQL or Cloud Memorystore for Redis) are authenticated via [Workload Identity](https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity).

## Override Brokerpak defaults

Brokerpaks are essentially a Terraform plan and related dependencies in a tar
file. You can inspect the Terraform plans to see what the defaults are, and then
you can tell Kf Cloud Service Broker to override them when creating new services.

For example, the [Terraform plan for MySQL](https://github.com/cloudfoundry-incubator/csb-brokerpak-gcp/blob/main/terraform/cloud-sql-provision.tf) includes a variable called `authorized_network`. If not overridden, the `default` VPC will be used. If you'd like to override the default, you can pass that during service creation. Here are some examples:

1. Override the compute region `config`.

  ```sh
   kf create-service csb-google-postgres small spring-music-postgres-db -c '{"config":"YOUR_COMPUTE_REGION"}'
  ```

1. Override the `authorized_network` and compute region `config`.

  ```sh
   kf create-service csb-google-postgres small spring-music-postgres-db -c '{"config":"YOUR_COMPUTE_REGION","authorized_network":"YOUR_CUSTOM_VPC_NAME"}'
  ```

You can learn more by reading the [MySQL Plans and Configs](https://github.com/cloudfoundry-incubator/csb-brokerpak-gcp/blob/main/docs/mysql-plans-and-config.md) documentation.

## Architecture

The following Kf Cloud Service Broker architecture shows how instances are created.

{{<figure src="./kf-csb-architecture.svg" alt="Kf Cloud Service Broker architecture">}}

*   The Kf Cloud Service Broker (CSB) is installed in its own namespace.
*   On installation, a MySQL instance must be provided to persist
    business logic used by Kf Cloud Service Broker. Requests are sent securely
    from the Kf Cloud Service Broker pod to the MySQL instance via
    the MySQL Auth Proxy.
*   On service provisioning, a Kf Service custom resource
    is created. The reconciler of the Kf Service
    provisions Google Cloud backing services using the Open Service Broker API.
*   When a request to provision/deprovision backing resources is received,
    Kf Cloud Service Broker sends resource creation/deletion requests to the
    corresponding Google Cloud service, and these requests are authenticated
    with Workload Identity. It also persists the business logics (e.g. mapping of
    Kf services to backing services, service bindings) to
    the MySQL instance.
*   On backing service creation success, the backing service is bound to an App
    via [VCAP_SERVICES]({{< relref "app-runtime#vcapservices" >}}).

## What's next?

*   [Deploy Kf Cloud Service Broker]({{< relref "deploying-cloud-sb" >}}).
*   [Learn how to list and provision services]({{< relref "managed-services" >}}).

{% endblock %}
