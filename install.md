
## Pre-requisites

This guide is intended to provide you with all the commands you'll
need to install `kf` in a single place. It assumes you have a the
ability to run root containers in a cluster with at least 12 vCPUs
and 45G of memory and a minimum of three nodes.

You will also need to provide a docker compatable registry.

## Configure your Registry
In order to make this install simple to walk through we recommend you
store your docker registry details in a an environment variable. This
install guide uses gcr on gke.

```
export GCP_PROJECT=<PROJECT_ID>
export KF_REGISTRY=gcr.io/$GCP_PROJECT
```

## Install Istio && Knative

Install istio CRDs and deploy pods and label the default namespace.
```
kubectl apply --filename https://github.com/knative/serving/releases/download/v0.5.0/istio-crds.yaml && \
kubectl apply --filename https://github.com/knative/serving/releases/download/v0.5.0/istio.yaml && \
kubectl label namespace default istio-injection=enabled
```

Install knative CRDs.
```
kubectl apply --selector knative.dev/crd-install=true \
--filename https://github.com/knative/serving/releases/download/v0.5.0/serving.yaml \
--filename https://github.com/knative/build/releases/download/v0.5.0/build.yaml \
--filename https://github.com/knative/eventing/releases/download/v0.5.0/release.yaml \
--filename https://github.com/knative/eventing-sources/releases/download/v0.5.0/eventing-sources.yaml \
--filename https://github.com/knative/serving/releases/download/v0.5.0/monitoring.yaml \ &&
--filename https://raw.githubusercontent.com/knative/serving/v0.5.0/third_party/config/build/clusterrole.yaml
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

If you want to go more in depth installing knative check out [their docs](knative).


## Upload buildpacks
Buildpacks are provided by the operator and can be uploaded to Knative using
a script. A set of buidpacks is included in this repo. They can be installed
with the following:

```
./hack/upload-buildpacks.sh
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
broker. You can use helm to install the gcp-service-broker from
the third_party directory.

Configure GCP service account & APIs
```
gcloud iam service-accounts create \
    gcp-service-broker

gcloud iam service-accounts keys \
    create /tmp/key.json --iam-account \
    gcp-service-broker@$GCP_PROJECT.iam.gserviceaccount.com

gcloud projects \
    add-iam-policy-binding \
    $GCP_PROJECT --member \
    serviceAccount:gcp-service-broker@$GCP_PROJECT.iam.gserviceaccount.com \
    --role "roles/owner"

gcloud services enable \
    cloudresourcemanager.googleapis.com \
    iam.googleapis.com
```

Once you have your key.json, copy this in the values.yaml file
in `/third_party/gcp-service-broker` file.


Configure helm
```
kubectl create serviceaccount --namespace kube-system tiller
kubectl create clusterrolebinding tiller-cluster-rule \
--clusterrole=cluster-admin --serviceaccount=kube-system:tiller
helm init --service-account tiller
```

Install the gcp-service broker
```
cd ./third_party/gcp-service-broker/
helm install . --name gcp-service-broker
```

[knative]: https://github.com/knative/docs/tree/master/docs/install
