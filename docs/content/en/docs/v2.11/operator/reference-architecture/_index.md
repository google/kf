---
title: "Kf reference architecture diagrams"
description: >
    Learn about common ways of deploying and managing Kf clusters.
---

Kf can benefit from the rich ecosystem of Anthos and GKE, including automation, managed backing services, and development tools.

## GKE reference architecture

Kf clusters are managed just like any other GKE cluster, and access resources in the same way.

{{< figure src="./gke-reference-diagram.svg" alt="GKE and Kf cluster architecture." >}}

## ConfigSync

Kf can work with
[ConfigSync](https://github.com/GoogleContainerTools/kpt-config-sync) to
automate Space creation across any number of clusters. 
Create new namespaces on non-Kf clusters, and Spaces on Kf clusters. 
You can also manage Spaces by removing the Kf CLI from Space management, if desired. Use `kubectl explain` to [get Kf CRD specs]({{< relref "kf-dependencies#get_crd_details" >}}) to fully manage your product configuration via GitOps.

{{< figure src="./config-sync-reference-diagram.svg" alt="Config Sync and Kf cluster architecture." >}}
