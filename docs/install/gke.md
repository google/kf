# Prepping GKE for KF

This guide walks you through the installation of GKE with Cloud Run with the
intent of using it with `kf`.

> Note: Installing Cloud Run is equivalent to installing Knative Serve and
> Istio. Therefore, only Knative Build is required after.

## Before you begin

This guide assumes you are using `bash` in a Mac or Linux environment; some
commands will need to be adjusted for use in a Windows environment.

### Install Cloud SDK

1. If you already have `gcloud` installed you can skip these steps.
1. Download and install the gcloud command line tool:
   https://cloud.google.com/sdk/install
1. Authorize `gcloud`:
   ```sh
   gcloud auth login
   ```

### Setup environment variables

We will use these environment variables to simplify the installation.

```
export CLUSTER_NAME=[REPLACE]
export ZONE=[REPLACE]
export PROJECT_ID=[REPLACE]
export SERVICE_ACCOUNT=$CLUSTER_NAME@$PROJECT_ID.iam.gserviceaccount.com
export NETWORK=projects/$PROJECT_ID/global/networks/$CLUSTER_NAME
```

*NOTE: Replace the `[REPLACE]` value with the according values.*

### Create a service account and give it `roles/storage.admin` on your project:

```sh
gcloud iam service-accounts create $CLUSTER_NAME --project $PROJECT_ID`
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member=serviceAccount:$SERVICE_ACCOUNT \
  --role="roles/storage.admin"
```

### Create a network

If you have an existing network you wish to use, customize the `NETWORK` env var set previously to point to your network and skip this step.

```sh
gcloud compute networks create $CLUSTER_NAME --project $PROJECT_ID
```

### Create the Kubernetes cluster

```sh
gcloud beta container clusters create $CLUSTER_NAME \
  --zone $ZONE \
  --no-enable-basic-auth \
  --cluster-version "1.13.6-gke.13" \
  --machine-type "n1-standard-1" \
  --image-type "COS" \
  --disk-type "pd-standard" \
  --disk-size "100" \
  --metadata disable-legacy-endpoints=true \
  --service-account $SERVICE_ACCOUNT \
  --num-nodes "3" \
  --enable-stackdriver-kubernetes \
  --enable-ip-alias \
  --network $NETWORK \
  --default-max-pods-per-node "110" \
  --addons HorizontalPodAutoscaling,HttpLoadBalancing,Istio,CloudRun \
  --istio-config auth=MTLS_PERMISSIVE \
  --enable-autoupgrade \
  --enable-autorepair \
  --project $PROJECT_ID
```

### Target your cluster:

```sh
gcloud container clusters get-credentials $CLUSTER_NAME --zone $ZONE --project $PROJECT_ID
```

### Next steps
Continue with the [install docs](/docs/install.md) to install Kf into the cluster you just created.
