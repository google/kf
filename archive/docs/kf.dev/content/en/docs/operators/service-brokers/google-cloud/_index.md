---
title: "Google Cloud Broker"
linkTitle: "Google Cloud"
weight: 20
description: >
  Learn how to install the Google Cloud service broker for use with Kf.
---

## Prerequisites

[kf]: /docs/getting-started/install
In addition to a Kubernetes cluster with Kf and Service Catalog installed (see
[these instructions][kf]), the following tools must be installed on the
workstation where you will be using the `kf` CLI:

[gcloud]: https://cloud.google.com/sdk/install
[helm]: https://helm.sh/docs/using_helm/
1. **`gcloud`**: Follow [these instructions][gcloud] to install the `gcloud`
   CLI.
1. **`helm`**: Follow [these instructions][helm] to install the `helm`
   CLI.
1. **`git`**: Git is required to clone a repository.

## GCP Configuration

### Select a GCP project

Choose the project you want to use with the rest of this tutorial. You must be
an owner of the project you choose. Set these environment variables in your terminal:

```sh
export GOOGLE_PROJECT=<replace-with-your-project-id>
export SERVICE_ACCOUNT_NAME=kf-gcp-broker
```

### Create a service account for the broker

1. Create the service account:

    ```sh
    gcloud iam service-accounts create $SERVICE_ACCOUNT_NAME
	```

1. Create new credentials to let the broker authenticate, and download the credential
to `key.json`:

    ```sh
    gcloud iam service-accounts keys create key.json --iam-account $SERVICE_ACCOUNT_NAME@$GOOGLE_PROJECT.iam.gserviceaccount.com
	```

1. Grant project owner permissions to the broker:

	```sh
    gcloud projects add-iam-policy-binding $GOOGLE_PROJECT --member serviceAccount:$SERVICE_ACCOUNT_NAME@$GOOGLE_PROJECT.iam.gserviceaccount.com --role "roles/owner"
	```

### Enable required APIs

Now you need to enable APIs to let the broker provision resources.

The broker has a few APIs that are required for it to run, and a few that are
optional but must be enabled to provision resources of a particular type.

The Cloud Resource Manager and IAM APIs are required for the broker to run. Enable them:

```sh
gcloud services enable cloudresourcemanager.googleapis.com iam.googleapis.com --project $GOOGLE_PROJECT
```

### Enable service APIs

The following APIs must be enabled to use their respective services.
For example, you must enable the BigQuery API on the project if you want to
provision and use BigQuery instances.

1. [BigQuery API](https://console.cloud.google.com/apis/api/bigquery/overview)
1. [BigTable API](https://console.cloud.google.com/apis/api/bigtableadmin/overview)
1. [CloudSQL API](https://console.cloud.google.com/apis/library/sql-component.googleapis.com)
1. [CloudSQL Admin API](https://console.cloud.google.com/apis/library/sqladmin.googleapis.com)
1. [Datastore API](https://console.cloud.google.com/apis/api/datastore.googleapis.com/overview)
1. [Pub/Sub API](https://console.cloud.google.com/apis/api/pubsub/overview)
1. [Redis API](https://console.cloud.google.com/apis/api/redis.googleapis.com/overview)
1. [Storage API](https://console.cloud.google.com/apis/api/storage_component/overview)
1. [Spanner API](https://console.cloud.google.com/apis/api/spanner/overview)

You can always enable the APIs later. If you try to provision an instance that
uses a disabled API then the provisioning will fail.

## Helm

Run the following commands to configure Helm in your cluster:

```sh
kubectl create serviceaccount --namespace kube-system tiller
kubectl create clusterrolebinding tiller-cluster-rule \
  --clusterrole=cluster-admin --serviceaccount=kube-system:tiller
helm init --service-account tiller
```

## Service Broker

1. Clone the service broker to your local workstation and `cd` to it:

    ```sh
    git clone https://github.com/google/kf.git
    cd kf/third_party/gcp-service-broker 
    ```

1. Update the dependencies of the Helm chart:

	```sh
    helm dependency update
	```

    **Optional:** read through the rest of the properties and change any you need
     to fit your environment.
   
1. Copy the contents of the service account key in `key.json` to the
`broker.service_account_json` value in `values.yaml`.

1. Install the broker:

	```sh
    helm install --name kf-google-broker .
	```

## Confirm broker installation

It may take several minutes for your broker to come up and register itself with Service Catalog.

Run `kf marketplace` until you see output similar to:

```sh
BROKER              NAME                           NAMESPACE  STATUS  DESCRIPTION
gcp-service-broker  google-stackdriver-profiler               Active  Continuous CPU and heap profiling to improve performance and reduce costs.
gcp-service-broker  google-stackdriver-monitoring             Active  Stackdriver Monitoring provides visibility into the performance, uptime, and overall health of cloud
gcp-service-broker  google-dataflow                           Active  A managed service for executing a wide variety of data processing patterns built on Apache Beam.
gcp-service-broker  google-cloudsql-mysql                     Active  Google CloudSQL for MySQL is a fully-managed MySQL database service.
gcp-service-broker  google-spanner                            Active  The first horizontally scalable, globally consistent, relational database service.
gcp-service-broker  google-ml-apis                            Active  Machine Learning APIs including Vision, Translate, Speech, and Natural Language.
gcp-service-broker  google-pubsub                             Active  A global service for real-time and reliable messaging and streaming data.
gcp-service-broker  google-datastore                          Active  Google Cloud Datastore is a NoSQL document database service.
gcp-service-broker  google-stackdriver-debugger               Active  Stackdriver Debugger is a feature of the Google Cloud Platform that lets you inspect the state of an
gcp-service-broker  google-firestore                          Active  Cloud Firestore is a fast, fully managed, serverless, cloud-native NoSQL document database that simp
gcp-service-broker  google-bigtable                           Active  A high performance NoSQL database service for large analytical and operational workloads.
gcp-service-broker  google-storage                            Active  Unified object storage for developers and enterprises. Cloud Storage allows world-wide storage and r
gcp-service-broker  google-stackdriver-trace                  Active  Stackdriver Trace is a distributed tracing system that collects latency data from your applications
gcp-service-broker  google-cloudsql-postgres                  Active  Google CloudSQL for PostgreSQL is a fully-managed PostgreSQL database service.
gcp-service-broker  google-dialogflow                         Active  Dialogflow is an end-to-end, build-once deploy-everywhere development suite for creating conversatio
gcp-service-broker  google-bigquery                           Active  A fast, economical and fully managed data warehouse for large-scale data analytics.
```

{{% alert title="Ready!" color="primary" %}}
The Google Service Broker is installed and can be used to create services and bind them
to apps you deploy with Kf.
{{% /alert %}}
