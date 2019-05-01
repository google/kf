# Install the Service Broker into Kubernetes for use with Cloud Foundry

## Introduction

This tutorial will walk you through installing the GCP Service Broker into
Kubernetes using the Helm release and using it with Cloud Foundry.

<walkthrough-tutorial-duration duration="15">
</walkthrough-tutorial-duration>

**Prerequisites**:

* A Kubernetes cluster you want to install the broker into.
* The `cf` and `helm` CLI tools.

## Select a project

Choose the project you want to use with the rest of this tutorial.

The project MUST have a Kubernetes cluster in it to deploy the broker into and
you MUST be an owner of the project.

<walkthrough-project-setup>
</walkthrough-project-setup>

## Create a service account for the broker

<walkthrough-watcher-constant key="service-account-name" value="gcp-service-broker">
</walkthrough-watcher-constant>


Create the service account:

    gcloud iam service-accounts create {{service-account-name}}

Create new credentials to let the broker authenticate:

    gcloud iam service-accounts keys create key.json --iam-account {{service-account-name}}@{{project-id}}.iam.gserviceaccount.com

Grant project owner permissions to the broker:

    gcloud projects add-iam-policy-binding {{project-id}} --member serviceAccount:{{service-account-name}}@{{project-id}}.iam.gserviceaccount.com --role "roles/owner"

## Enable required APIs

Now you need to enable APIs to let the broker provision those kind of resources.

The broker has a few APIs that are required for it to run, and a few that are
optional but must be enabled to provision resources of a particular type.

Enable the Cloud Resource Manager and IAM APIs to allow the service broker to run:

    gcloud services enable cloudresourcemanager.googleapis.com iam.googleapis.com

### Enable service APIs

The following APIs must be enabled to use their respective services.
For example, you must enable the BigQuery API on the project if you want to
provision and use BigQuery instances.

1. [BigQuery API](https://console.cloud.google.com/apis/api/bigquery/overview)
1. [BigTable API](https://console.cloud.google.com/apis/api/bigtableadmin/overview)
1. [CloudSQL API](https://console.cloud.google.com/apis/library/sql-component.googleapis.com)
1. [Datastore API](https://console.cloud.google.com/apis/api/datastore.googleapis.com/overview)
1. [Pub/Sub API](https://console.cloud.google.com/apis/api/pubsub/overview)
1. [Redis API](https://console.cloud.google.com/apis/api/redis.googleapis.com/overview)
1. [Storage API](https://console.cloud.google.com/apis/api/storage_component/overview)
1. [Spanner API](https://console.cloud.google.com/apis/api/spanner/overview)

You can always enable the APIs later. If you try to provision an instance that
uses a disabled API then the provisioning will fail.

## Install the broker

First, update the dependencies of the helm chart:

    helm dependency update

**Optional:** read through the rest of the properties and change any you need
   to fit your environment.
   
Next, copy the contents of the service account key in `key.json` 
to the `broker.service_account_json` value in `values.yaml`.

Finally, install the broker:

    helm install --name gsb-tutorial --set svccat.register=false .

## Install the broker into Cloud Foundry

Follow the notes output by the previous command to get the credentials from your
Kubernetes cluster and install the broker.

If you need to see the notes again run the command:

    helm get notes gsb-tutorial

## Enable services for your developers

You can look at the installed service offerings in Cloud Foundry, you must make
these available to users before they can be used:

    cf service-offerings

List who has access to each service:

    cf service-access

Run `cf enable-service-access SERVICENAME` for each service you want to enable:

  cf enable-service-access google-storage

For more information, see the Cloud Foundry docs on [managing Service Brokers](https://docs.cloudfoundry.org/services/managing-service-brokers.html).
