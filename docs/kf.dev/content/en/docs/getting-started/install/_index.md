---
title: "Install Kf"
linkTitle: "Basic Kf Installation"
weight: 10
description: >
  Learn how to use the `kf` CLI to create a new Kubernetes cluster with Kf
  installed.
---

[for-operators]: /docs/operators
[gke]: https://cloud.google.com/kubernetes-engine/
Follow these instructions to quickly configure and deploy a basic Kf
installation on a new [Google Kubernetes Engine (GKE)][gke] cluster. For a more
comprehensive installation guide for system operators/admins, please see the
[Operators Guide][for-operators].

## Prerequisites

### Google Cloud
[sign-up]: https://cloud.google.com
[free]: https://cloud.google.com/free
[create-billing]: https://cloud.google.com/billing/docs/how-to/manage-billing-account

1. If you do not have a GCP account, you can [sign up][sign-up] (and you may be eligible for [free credit][free]).
2. If you do not have a GCP project, the installer can create one for you but you must
first configure a billing account as described in [these instructions][create-billing].

### Workstation

The following tools must be installed on the workstation where you will be using
the `kf` CLI:

[gcloud]: https://cloud.google.com/sdk/install
1. **`gcloud`**: Follow [these instructions][gcloud] to install the `gcloud`
   CLI.
1. **`kubectl`**: If you do not have `kubectl` installed, run `gcloud components install kubectl` to install it.

## Download the CLI

The `kf` CLI is built nightly from the master branch. It can be downloaded
from the following URLs:

### Linux
> https://storage.googleapis.com/artifacts.kf-releases.appspot.com/nightly/latest/bin/kf-linux
```sh
wget https://storage.googleapis.com/artifacts.kf-releases.appspot.com/nightly/latest/bin/kf-linux -O kf
chmod +x kf
sudo mv kf /usr/local/bin
```

### Mac
> https://storage.googleapis.com/artifacts.kf-releases.appspot.com/nightly/latest/bin/kf-darwin
```sh
curl https://storage.googleapis.com/artifacts.kf-releases.appspot.com/nightly/latest/bin/kf-darwin --output kf
chmod +x kf
sudo mv kf /usr/local/bin
```

### Windows
> https://storage.googleapis.com/artifacts.kf-releases.appspot.com/nightly/latest/bin/kf-windows.exe

## Create a cluster and install Kf

Run `kf install gke` in your terminal to begin the interactive installation.

{{% alert title="Important" color="secondary" %}}
* You are strongly encouraged to create a new cluster rather than use an existing
cluster.
* The installer can use an existing GCP project or create a new one.
{{% /alert %}}

Output will look similar to this:

```sh
$ kf install gke
kubectl version:
Client Version: v1.15.2
gcloud version:
Google Cloud SDK 257.0.0
alpha 2019.08.02
beta 2019.08.02
bq 2.0.46
core 2019.08.02
gsutil 4.41
kubectl 2019.08.02
[Select Project] finding your projects...
✔ kf-install-instructions
[Select Cluster] finding your GKE clusters...
✔ Create New GKE Cluster
[Ensure Billing] checking if kf-install-instructions has billing enabled
[Create New GKE Config] enabling required service APIs
✔ yes
[Enable Service API] enabling compute.googleapis.com service API. This may take
a moment
✔ yes
[Enable Service API] enabling container.googleapis.com service API. This may
take a moment
Cluster Name: kf-bw64yfs24o22
✔ yes
[Create Service Account] Creating service account
kf-bw650qyo0owr@kf-install-instructions.iam.gserviceaccount.com
```

### Confirm installation
After `kf install gke` has completed, you can confirm that your cluster is ready
by listing the spaces available in Kf. There should be one space that was
created by the installer:

```sh
 kf spaces
 Name               Age     Ready   Reason
 space-bw6548xzu4i8 2h      True
 ```

### Hello, World!
You can quickly deploy a simple "Hello, World!" sample app to further validate
your configuration. Simply run these commands to do that:

```sh
git clone https://github.com/google/kf
cd kf/samples/apps/helloworld
kf push helloworld
```

Confirm the app is running:

```sh
kf apps
```

Once the app is running, open a proxy to your app and navigate to
http://localhost:8080 in your browser:

```sh
kf proxy helloworld
```

{{% alert title="Ready!" color="primary" %}}
[spring-music]: /docs/getting-started/spring-music
Kf is installed and you're ready to deploy an application. Check out the [Spring
Music Getting Started][spring-music] documentation.
{{% /alert %}}
