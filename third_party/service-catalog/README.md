# Service Catalog Installation

## Install
To install Service Catalog in your cluster without Helm or Tiller, run:

```
kubectl apply --recursive --filename ./manifests/catalog
```

## About
This directory contains the rendered Service Catalog Helm chart to eliminate the need run Helm and Tiller in your cluster. To regenerate the chart, edit `values/catalog.yml`, then run:

```
helm template \
  --values ./values/prometheus.yaml \
  --output-dir ./manifests \
    ./charts/prometheus
```
