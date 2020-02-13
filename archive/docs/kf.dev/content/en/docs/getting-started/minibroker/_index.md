---
title: "Install a Basic Service Broker"
linkTitle: "Install a Basic Service Broker"
weight: 20 
description: >
  Learn how to install a basic service broker to provide database services to
  your applications in Kf.
---

[minibroker]: https://github.com/kubernetes-sigs/minibroker
Most Cloud Foundry service brokers comply with the OSB specification.
The following steps will guide you through installing a broker called
[Minibroker][minibroker] which will deploy services as Helm charts directly into your cluster.

It provides the following services:

* MariaDB
* MongoDB
* MySQL
* PostgreSQL
* Redis

Configure helm in your cluster:

```sh
kubectl create serviceaccount --namespace kube-system tiller
kubectl create clusterrolebinding tiller-cluster-rule \
--clusterrole=cluster-admin --serviceaccount=kube-system:tiller
helm init --service-account tiller
```

Add the chart and install:

```sh
helm repo add minibroker https://minibroker.blob.core.windows.net/charts
helm install --name minibroker --namespace minibroker minibroker/minibroker
```

It will take a while to start and register itself, after it's done you can
run kf marketplace again to see the services:

```sh
$ kf marketplace
5 services can be used in namespace "default", use the --service flag to list the plans for a service

BROKER              NAME                           NAMESPACE  STATUS  DESCRIPTION
minibroker          mariadb                                   Active  Helm Chart for mariadb
minibroker          mongodb                                   Active  Helm Chart for mongodb
minibroker          mysql                                     Active  Helm Chart for mysql
minibroker          postgresql                                Active  Helm Chart for postgresql
minibroker          redis                                     Active  Helm Chart for redis
```
