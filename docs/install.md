
## Pre-requisites

This guide is intended to provide you with all the commands you'll
need to install `kf` into an existing Kubernetes cluster.

It assumes you have:

* A Kubernetes cluster that:
  * Can run containers as root.
  * Has at least 12 vCPUs.
  * Has at least 45G of memory.
  * Has a minimum of three nodes.
* A Docker compatible container registry that you can write to.

## Configure your registry

In order to make this install simple to walk through we recommend you
store your Docker registry details in an environment variable. This
install guide uses Google Container Registry (GCR) on GKE.

```
export KF_REGISTRY=<your-container-registry>
e.g: export KF_REGISTRY=gcr.io/<PROJECT_ID>
```

## Install dependencies

`kf` uses Istio to route HTTP requests to the running applications and Knative
to deploy and scale applications.

Install Istio:

```.sh
kubectl apply --filename https://raw.githubusercontent.com/knative/serving/v0.6.1/third_party/istio-1.1.3/istio-crds.yaml && \
kubectl apply --filename https://raw.githubusercontent.com/knative/serving/v0.6.1/third_party/istio-1.1.3/istio.yaml && \
kubectl label namespace default istio-injection=enabled
```

Install Knative:

```.sh
kubectl apply --filename https://github.com/knative/serving/releases/download/v0.6.1/serving.yaml \
--filename https://github.com/knative/build/releases/download/v0.6.0/build.yaml \
--filename https://github.com/knative/serving/releases/download/v0.6.1/monitoring.yaml \
--filename https://raw.githubusercontent.com/knative/serving/v0.6.1/third_party/config/build/clusterrole.yaml
```

If you want to go more in depth installing Knative check out [their docs][knative].

Install the service catalog from the `third_party` directory included in this repo:

```.sh
kubectl apply -R -f third_party/service-catalog/manifests/catalog/templates
```


## Upload buildpacks

Buildpacks are provided by the operator and can be uploaded to Knative using
a script. A set of buidpacks is included in this repo. They can be installed
with the following:

```.sh
./hack/upload-buildpacks.sh
```

## Install the service catalog

You can install the service catalog from the `third_party` directory included
in this repo:

```.sh
kubectl apply -R -f third_party/service-catalog/manifests/catalog/templates
```

You should be able to see an empty marketplace at this point by running.

```.sh
kf marketplace
```

## Test your installation

At this point, your installation is set up and ready for use with `kf`.

Run `kf doctor` to validate it. You should see output like the following:

```
=== RUN	doctor/cluster
--- PASS: doctor/cluster
    --- PASS: doctor/cluster/Version
    --- PASS: doctor/cluster/Components
        --- PASS: doctor/cluster/Components/Kubernetes V1
            --- PASS: doctor/cluster/Components/Kubernetes V1/configmaps
            --- PASS: doctor/cluster/Components/Kubernetes V1/secrets
        --- PASS: doctor/cluster/Components/Knative Serving
            --- PASS: doctor/cluster/Components/Knative Serving/configurations
            --- PASS: doctor/cluster/Components/Knative Serving/routes
            --- PASS: doctor/cluster/Components/Knative Serving/revisions
            --- PASS: doctor/cluster/Components/Knative Serving/services
        --- PASS: doctor/cluster/Components/Service Catalog
            --- PASS: doctor/cluster/Components/Service Catalog/clusterservicebrokers
=== RUN	doctor/buildpacks
=== RUN	doctor/buildpacks/Buildpacks
--- PASS: doctor/buildpacks
    --- PASS: doctor/buildpacks/Buildpacks
PASS
```

If the result is a failure, re-run the commands in the previous sections.

## Push your first app

Now you can deploy your first app using `kf`.
Run the following command to push it:

```.sh
kf push helloworld --container-registry $KF_REGISTRY
```

## (Optional) Install a service broker

You can install [Open Service Broker](https://www.openservicebrokerapi.org/)
compatible service brokers into your cluster now to allow users to create and
bind services.

You should be able to see an empty marketplace at this point by running.

```.sh
kf marketplace
```

Most Cloud Foundry service brokers comply with the OSB specification.
The following steps will guide you through installing a broker called
"mini-broker" which will deploy services as helm charts directly into your cluster.

It provides the following services:

* MariaDB
* MongoDB
* MySQL
* PostgreSQL
* Redis

Configure helm in your cluster:

```.sh
kubectl create serviceaccount --namespace kube-system tiller
kubectl create clusterrolebinding tiller-cluster-rule \
--clusterrole=cluster-admin --serviceaccount=kube-system:tiller
helm init --service-account tiller
```

Add the chart and install:

```.sh
helm repo add minibroker https://minibroker.blob.core.windows.net/charts
helm install --name minibroker --namespace minibroker minibroker/minibroker
```

It will take a while to start and register itself, after it's done you can
run kf marketplace again to see the services:

```
$ kf marketplace
5 services can be used in namespace "default", use the --service flag to list the plans for a service

BROKER              NAME                           NAMESPACE  STATUS  DESCRIPTION
minibroker          mariadb                                   Active  Helm Chart for mariadb
minibroker          mongodb                                   Active  Helm Chart for mongodb
minibroker          mysql                                     Active  Helm Chart for mysql
minibroker          postgresql                                Active  Helm Chart for postgresql
minibroker          redis                                     Active  Helm Chart for redis
```

[knative]: https://github.com/knative/docs/tree/master/docs/install
