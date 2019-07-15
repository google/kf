Prepping GKE for KF
===================

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

## Create a Kubernetes cluster with Cloud Run

> [Cloud Run on GKE](https://cloud.google.com/run/docs/gke/setup) is a hosted
> offering on top of GKE that builds around Istio and Knative Serving.

### Setup environment variables

We will use these environment variables to simplify the installation.

```
export CLUSTER_NAME=[REPLACE]
export ZONE=[REPLACE]
```

*NOTE: Replace the `[REPLACE]` value with the according values.*

### Create the Kubernetes cluster

```sh
gcloud beta container clusters create $CLUSTER_NAME \
--addons=HorizontalPodAutoscaling,HttpLoadBalancing,Istio,CloudRun \
--machine-type=n1-standard-4 \
--cluster-version=latest \
--zone=$ZONE \
--enable-stackdriver-kubernetes \
--enable-ip-alias \
--scopes cloud-platform
```
