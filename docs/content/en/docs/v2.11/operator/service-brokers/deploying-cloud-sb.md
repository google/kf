---
title: "Deploy Kf Cloud Service Broker"
---

This page shows you how to deploy Kf Cloud Service Broker and use it to provision or deprovision backing resources. 
Read about the [concepts and architecture]({{< relref "cloud-sb-overview" >}}) to learn more about the Kf Cloud Service Broker.

## Create environment variables {#create_env_variables}

* **Linux**

  <pre class="devsite-click-to-copy" translate="no">
  export PROJECT_ID=<var>YOUR_PROJECT_ID</var>
  export CLUSTER_PROJECT_ID=<var>YOUR_PROJECT_ID</var>
  export CLUSTER_NAME=<var>kf-cluster</var>
  export INSTANCE_NAME=<var>cloud-service-broker</var>
  export COMPUTE_REGION=<var>us-central1</var>
  </pre>

* **Windows PowerShell**

  <pre class="devsite-click-to-copy" translate="no">
  Set-Variable -Name PROJECT_ID -Value <var>YOUR_PROJECT_ID</var>
  Set-Variable -Name CLUSTER_PROJECT_ID -Value <var>YOUR_PROJECT_ID</var>
  Set-Variable -Name CLUSTER_NAME -Value <var>kf-cluster</var>
  Set-Variable -Name INSTANCE_NAME -Value <var>cloud-service-broker</var>
  Set-Variable -Name COMPUTE_REGION -Value <var>us-central1</var>
  </pre>

## Set up the Kf Cloud Service Broker database

1. Create a MySQL instance.

    {{%note%}}
    Read [Creating and managing MySQL users](https://cloud.google.com/sql/docs/mysql/create-manage-users) for Google Cloud SQL and set a secure password for the default `root` user.
    {{%/note%}}

    ```sh
    gcloud sql instances create ${INSTANCE_NAME} --cpu=2 --memory=7680MB --require-ssl --region=${COMPUTE_REGION}
    ```

1. Create a database named `servicebroker` in the MySQL instance.

   {{%note%}}Document the database name since it is used in later steps.{{%/note%}}

   <pre class="devsite-click-to-copy" translate="no">
   gcloud sql databases create servicebroker -i ${INSTANCE_NAME}</pre>

1. Create a username and password to be used by the broker.

   {{%note%}}Document these values since they will be used in later steps.{{%/note%}}

   <pre class="devsite-click-to-copy" translate="no">
   gcloud sql users create <var>csbuser</var> -i ${INSTANCE_NAME} --password=<var>csbpassword</var></pre>

## Set up a Google Service Account for the broker

1. Create a Google Service Account.

  <pre class="devsite-click-to-copy" translate="no">
  gcloud iam service-accounts create csb-${CLUSTER_NAME}-sa \
      --project=${CLUSTER_PROJECT_ID} \
      --description="GSA for CSB at ${CLUSTER_NAME}" \
      --display-name="csb-${CLUSTER_NAME}"</pre>

1. Grant `roles/cloudsql.client` permissions to the Service Account. This is required to connect the service broker pod to the CloudSQL for MySQL instance through the CloudSQL Proxy.

  <pre class="devsite-click-to-copy" translate="no">
  gcloud projects add-iam-policy-binding ${CLUSTER_PROJECT_ID} \
      --member="serviceAccount:csb-${CLUSTER_NAME}-sa@${CLUSTER_PROJECT_ID}.iam.gserviceaccount.com" \
      --role="roles/cloudsql.client"</pre>

1. Grant additional Google Cloud permissions to the Service Account.

  {{%note%}}
  In the example below, we grant IAM roles required to provision an instance of CloudSQL and Cloud Memorystore.
  You must grant this service account the appropriate roles to provision instances of other Google Cloud managed services listed in `kf marketplace`.
  {{%/note%}}

  <pre class="devsite-click-to-copy" translate="no">
  gcloud projects add-iam-policy-binding ${CLUSTER_PROJECT_ID} \
      --member="serviceAccount:csb-${CLUSTER_NAME}-sa@${CLUSTER_PROJECT_ID}.iam.gserviceaccount.com" \
      --role="roles/compute.networkUser"</pre>

  <pre class="devsite-click-to-copy" translate="no">
  gcloud projects add-iam-policy-binding ${CLUSTER_PROJECT_ID} \
      --member="serviceAccount:csb-${CLUSTER_NAME}-sa@${CLUSTER_PROJECT_ID}.iam.gserviceaccount.com" \
      --role="roles/cloudsql.admin"</pre>

  <pre class="devsite-click-to-copy" translate="no">
  gcloud projects add-iam-policy-binding ${CLUSTER_PROJECT_ID} \
      --member="serviceAccount:csb-${CLUSTER_NAME}-sa@${CLUSTER_PROJECT_ID}.iam.gserviceaccount.com" \
      --role="roles/redis.admin"</pre>

1. Verify the permissions.

  {{%warning%}}
  Replace the `CSB_SERVICE_ACCOUNT_NAME` variable in the YAML below with the full service account resolved from `csb-${CLUSTER_NAME}-sa@${CLUSTER_PROJECT_ID}.iam.gserviceaccount.com`.
  {{%/warning%}}

  <pre class="devsite-click-to-copy" translate="no">
  gcloud projects get-iam-policy ${CLUSTER_PROJECT_ID} \
      --filter='bindings.members:serviceAccount:"<var>CSB_SERVICE_ACCOUNT_NAME</var>"' \
      --flatten="bindings[].members"</pre>

## Set up Workload Identity for the broker

1. Bind the Google Service Account with the Kubernetes Service Account.

  <pre class="devsite-click-to-copy" translate="no">
  gcloud iam service-accounts add-iam-policy-binding "csb-${CLUSTER_NAME}-sa@${CLUSTER_PROJECT_ID}.iam.gserviceaccount.com" \
      --project=${CLUSTER_PROJECT_ID} \
      --role="roles/iam.workloadIdentityUser" \
      --member="serviceAccount:${CLUSTER_PROJECT_ID}.svc.id.goog[kf-csb/csb-user]"</pre>

1. Verify the binding.

  <pre class="devsite-click-to-copy" translate="no">
  gcloud iam service-accounts get-iam-policy "csb-${CLUSTER_NAME}-sa@${CLUSTER_PROJECT_ID}.iam.gserviceaccount.com" \
      --project=${CLUSTER_PROJECT_ID}</pre>

## Set up a Kubernetes Secret to share configuration with the broker {#kubernetes_secret}

1. Create a config.yml file.

  Note: Replace the default user/password if desired. Ensure you have set the `CLUSTER_PROJECT_ID` in the [Create environment variables](#create_env_variables) step.

  ```sh
  cat << EOF >> ./config.yml
  gcp:
    credentials: ""
    project: ${CLUSTER_PROJECT_ID}

  db:
    host: 127.0.0.1
    password: csbpassword
    user: csbuser
    tls: false
  api:
    user: servicebroker
    password: password
  EOF
  ```

1. Create the `kf-csb` namespace.

  ```sh
  kubectl create ns kf-csb
  ```

1. Create the Kubernetes Secret.

  ```sh
  kubectl create secret generic csb-secret --from-file=config.yml -n kf-csb
  ```

## Install the Kf Cloud Service Broker

1. Download the `kf-csb.yml`.

  ```sh
  gsutil cp gs://kf-releases/csb/v{{this_kf_csb_version}}/kf-csb.yaml /tmp/kf-csb.yaml
  ```

1. Edit `/tmp/kf-csb.yaml` and replace placeholders with final values. In the example below, `sed` is used.

  ```sh
  sed -i "s|<GSA_NAME>|csb-${CLUSTER_NAME}-sa@${CLUSTER_PROJECT_ID}.iam.gserviceaccount.com|g" /tmp/kf-csb.yaml

  sed -i "s|<INSTANCE_CONNECTION_NAME>|${CLUSTER_PROJECT_ID}:${COMPUTE_REGION}:${INSTANCE_NAME}|g" /tmp/kf-csb.yaml

  sed -i "s|<DB_PORT>|3306|g" /tmp/kf-csb.yaml
  ```

1. Apply yaml for Kf Cloud Service Broker.

  ```sh
  kubectl apply -f /tmp/kf-csb.yaml
  ```

1. Verify the Cloud Service Broker installation status.

  ```sh
  kubectl get pods -n kf-csb
  ```

## Create a service broker

  {{%note%}}
  The user/password must match what you entered in the [Kubernetes secret](#kubernetes_secret) step earlier.
  {{%/note%}}

  ```sh
  kf create-service-broker cloud-service-broker servicebroker password http://csb-controller.kf-csb/
  ```

## Validate installation

  Check for available services in the marketplace.

  ```sh
  kf marketplace
  ```

  If everything is installed and configured correctly, you should see the following:

  ```none
  $ kf marketplace

  Broker                Name                          Namespace  Description
  cloud-service-broker  csb-google-bigquery                      A fast, economical and fully managed data warehouse for large-scale data analytics.
  cloud-service-broker  csb-google-dataproc                      Dataproc is a fully-managed service for running Apache Spark and Apache Hadoop clusters in a simpler, more cost-efficient way.
  cloud-service-broker  csb-google-mysql                         Mysql is a fully managed service for the Google Cloud Platform.
  cloud-service-broker  csb-google-postgres                      PostgreSQL is a fully managed service for the Google Cloud Platform.
  cloud-service-broker  csb-google-redis                         Cloud Memorystore for Redis is a fully managed Redis service for the Google Cloud Platform.
  cloud-service-broker  csb-google-spanner                       Fully managed, scalable, relational database service for regional and global application data.
  cloud-service-broker  csb-google-stackdriver-trace             Distributed tracing service
  cloud-service-broker  csb-google-storage-bucket                Google Cloud Storage that uses the Terraform back-end and grants service accounts IAM permissions directly on the bucket.
  ```

## Clean up

1. Delete cloud-service-broker.

  ```sh
  kf delete-service-broker cloud-service-broker
  ```

1. Delete CSB components.

  ```sh
  kubectl delete ns kf-csb
  ```

1. Delete the broker's database instance.

    <pre class="devsite-click-to-copy" translate="no">
    gcloud sql instances delete ${INSTANCE_NAME} --project=${CLUSTER_PROJECT_ID}</pre>

1. Remove the IAM policy bindings.

    <pre class="devsite-click-to-copy" translate="no">
    gcloud projects remove-iam-policy-binding ${CLUSTER_PROJECT_ID} \
    --member='serviceAccount:csb-${CLUSTER_NAME}-sa@${CLUSTER_PROJECT_ID}.iam.gserviceaccount.com' \
    --role=roles/cloudsql.client</pre>

    <pre class="devsite-click-to-copy" translate="no">
    gcloud projects remove-iam-policy-binding ${CLUSTER_PROJECT_ID} \
    --member='serviceAccount:csb-${CLUSTER_NAME}-sa@${CLUSTER_PROJECT_ID}.iam.gserviceaccount.com' \
    --role=roles/compute.networkUser</pre>

    <pre class="devsite-click-to-copy" translate="no">
    gcloud projects remove-iam-policy-binding ${CLUSTER_PROJECT_ID} \
    --member='serviceAccount:csb-${CLUSTER_NAME}-sa@${CLUSTER_PROJECT_ID}.iam.gserviceaccount.com' \
    --role=roles/redis.admin</pre>

1. Remove the GSA.

    <pre class="devsite-click-to-copy" translate="no">
    gcloud iam service-accounts delete csb-${CLUSTER_NAME}-sa@${CLUSTER_PROJECT_ID}.iam.gserviceaccount.com \
      --project=${CLUSTER_PROJECT_ID}</pre>
