# Use an existing Kf cluster

To use an existing cluster, it is assumed your cluster has the following:

* Can run containers as root.
* Has at least 12 vCPUs.
* Has at least 45G of memory.
* Has a minimum of three nodes.

## Install dependencies

`kf` uses Istio to route HTTP requests to the running applications and Knative
to deploy and scale applications.

> Note: Installing Istio and Knative Serve can be skipped if you are [using
> Cloud Run on GKE](/docs/install/existing-cluster.md).

Install Istio:

```.sh
kubectl apply --filename https://raw.githubusercontent.com/knative/serving/v0.6.1/third_party/istio-1.1.3/istio-crds.yaml && \
kubectl apply --filename https://raw.githubusercontent.com/knative/serving/v0.6.1/third_party/istio-1.1.3/istio.yaml && \
kubectl label namespace default istio-injection=enabled
```

Install Knative Serve:

```.sh
kubectl apply --filename https://github.com/knative/serving/releases/download/v0.6.1/serving.yaml \
--filename https://github.com/knative/serving/releases/download/v0.6.1/monitoring.yaml \
--filename https://raw.githubusercontent.com/knative/serving/v0.6.1/third_party/config/build/clusterrole.yaml
```

## Confirm kubeconfig
The workstation you install `kf` on must have a valid `kubectl` configuration
located at `$HOME/.kube/config`.

### Next steps
Continue with the [install docs](docs/install.md) to install Kf into the cluster you just created.
