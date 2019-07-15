# Install Kf

## Pre-requisites

This guide is intended to provide you with all the commands you'll
need to install `kf` into an existing Kubernetes cluster.

It assumes you have:

* A Kubernetes cluster that:
  * Can run containers as root.
  * Has at least 12 vCPUs.
  * Has at least 45G of memory.
  * Has a minimum of three nodes.
* A Docker-compatible container registry that you can write to.

## Configure your registry

In order to make this install simple to walk through we recommend you
store your Docker registry details in an environment variable. This
install guide uses Google Container Registry (GCR) on GKE.

```
export KF_REGISTRY=gcr.io/<PROJECT_ID>
```

## Install `kf` CLI

The `kf` CLI is built nightly from the master branch. It can be downloaded
from the following URLs:

### Linux
> https://storage.googleapis.com/artifacts.kf-releases.appspot.com/nightly-builds/cli/kf-linux-latest
```sh
wget https://storage.googleapis.com/artifacts.kf-releases.appspot.com/nightly-builds/cli/kf-linux-latest -O kf
chmod +x kf
sudo mv kf /usr/local/bin
```

### Mac
> https://storage.googleapis.com/artifacts.kf-releases.appspot.com/nightly-builds/cli/kf-darwin-latest
```sh
wget https://storage.googleapis.com/artifacts.kf-releases.appspot.com/nightly-builds/cli/kf-darwin-latest -O kf
chmod +x kf
sudo mv kf /usr/local/bin
```

### Windows
> https://storage.googleapis.com/artifacts.kf-releases.appspot.com/nightly-builds/cli/kf-windows-latest.exe

## Create a Kubernetes cluster

* Google Cloud: [Create a GKE cluster](install/gke.md). Knative Serving and Istio will be installed with this cluster.

## Install dependencies

### Knative Build:

```.sh
kubectl apply --filename https://github.com/knative/build/releases/download/v0.6.0/build.yaml
```

> If you want more information about installing Knative, see [their docs][knative].

### Service Catalog:

```.sh
kubectl apply -R -f third_party/service-catalog/manifests/catalog/templates
```

## Install kf

Kf has controllers, reconcilers and webhooks that must be installed. The kf
containers and YAML are built nightly from the master branch. It can be
installed using the following:

```sh
kubectl apply -f https://storage.googleapis.com/artifacts.kf-releases.appspot.com/nightly-builds/releases/release-latest.yaml
```

You should be able to see an empty marketplace at this point by running:

```.sh
kf marketplace
```

## Test your installation

Your installation is set up and ready for use with `kf`.

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

## Create and target a space

```sh
kf create-space demo-space \
  --container-registry $KF_REGISTRY
kf target -s demo-space
```


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
