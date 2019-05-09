
## Pre-requisites

This guide is intended to provide you with all the commands you'll
need to install `kf` in a single place. It assumes you have a the 
ability to run root containers in a cluster with at least 12 vCPUs 
and 45G of memory and a minumum of three nodes.

You will also need to provide a docker compatable registry. 

## Configure your Registry
In order to make this install simple to walk through we recomend you 
store your docker registry details in a an environment variable. This
install guide uses gcr on gke. 

```
export KF_REGISTRY=gcr.io/<PROJECT_ID>
```

## Install Istio && Knative

Install istio CRDs and deploy pods and label the default namespace. 
```
kubectl apply --filename https://github.com/knative/serving/releases/download/v0.5.0/istio-crds.yaml && \
kubectl apply --filename https://github.com/knative/serving/releases/download/v0.5.0/istio.yaml && \
kubectl label namespace default istio-injection=enabled
```

Install Knative PODs
```
kubectl apply --filename https://github.com/knative/serving/releases/download/v0.5.0/serving.yaml \
--filename https://github.com/knative/build/releases/download/v0.5.0/build.yaml \
--filename https://github.com/knative/eventing/releases/download/v0.5.0/release.yaml \
--filename https://github.com/knative/eventing-sources/releases/download/v0.5.0/eventing-sources.yaml \
--filename https://github.com/knative/serving/releases/download/v0.5.0/monitoring.yaml \
--filename https://raw.githubusercontent.com/knative/serving/v0.5.0/third_party/config/build/clusterrole.yaml
```

If you want to go more in depth installing knative check out [thier docs](knative).


## Upload buildpacks
Buildpacks are provided by the operator and can be uploaded to Knative using the
CLI. A set of buidpacks is included in this repo. Change into the `buildpack-samples`
directory run the following command. 

```
kf upload-buildpacks --container-registry $KF_REGISTRY
```

## Push your first app
At this point you are ready to deploy your first app using `kf`. Run the following command 
to push your first app. 

```
kf push helloworld --container-registry $KF_REGISTRY
```

## Install the service catalog
You can install the service catalog from the third_party directory included 
in this repo. 

```
kubectl apply -R -f third_party/service-catalog/manifests/catalog/templates
```

You should be able to see an empty marketplace at this point by running.

```
kf marketplace
```

## Install a service broker
Once you have the service catalog you'll want to install a service
broker. This example uses a broker called "mini-broker" which will
deploy services as helm charts locally in you r cluster.

Configure helm in your cluster
```
kubectl create serviceaccount --namespace kube-system tiller 
kubectl create clusterrolebinding tiller-cluster-rule \
--clusterrole=cluster-admin --serviceaccount=kube-system:tiller
helm init --service-account tiller
```

Add the chart and install
```
helm repo add minibroker https://minibroker.blob.core.windows.net/charts
helm install --name minibroker --namespace minibroker minibroker/minibroker
```

[knative]: https://github.com/knative/docs/tree/master/docs/install