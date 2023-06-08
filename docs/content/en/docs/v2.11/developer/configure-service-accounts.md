---
title: "Configure and Use Service Accounts"
weight: 51
description: >
    Configure and Use Kubernetes Service Accounts from your Apps.
---

By default, all applications in Kf are assigned a unique Kubernetes Service Account (KSA) named `sa-<APP_NAME>`.
Kf uses this KSA as the "user" it runs application instances and tasks under.

Each App KSA receives a copy of the container registry credentials used by the Space's build KSA so Kf apps can pull
container images that were created during `kf push`.

## Using the service account

Kuberenetes Pods (the building blocks of Apps and Tasks) automatically receive a JWT for the KSA
mounted in the container:

```sh
$ ls /var/run/secrets/kubernetes.io/serviceaccount/
ca.crt
namespace
token
```

* `ca.crt` The Kubernetes control plane's certificate.
* `namespace` The Kubernetes namespace of the workload.
* `token` A Base64 encoded JWT for the Kf App's Service Account.

Below is an example of what the JWT looks like, note that:

* It expires and needs to be periodically refreshed from disk.
* It's audience is only valid within the Kubernetes cluster.

```json
{
    "aud": [
        "<KUBERNETES_CLUSTER_URI>"
    ],
    "exp": 3600,
    "iat": 0,
    "iss": "<KUBERNETES_CLUSTER_URI>",
    "kubernetes.io": {
        "namespace": "<SPACE_NAME>",
        "pod": {
            "name": "<APP_NAME>-<RANDOM_SUFFIX>",
            "uid": "<APP_GUID>"
        },
        "serviceaccount": {
            "name": "sa-<APP_NAME>",
            "uid": "<SERVICE_ACCOUNT_GUID>"
        },
        "warnafter": 3500
    },
    "nbf": 0,
    "sub": "system:serviceaccount:<SPACE_NAME>:sa-<APP_NAME>"
}
```

You can use this credential to connect to the Kubernetes control plane listed in the issuer (`iss`) field.

## Customizing the service account

{{< note >}}
[This feature must be enabled]({{<relref "customizing-features#ksa-overrides">}}) by your platform operator before being used.
{{< /note >}}

You want to use a different service account than the default one Kf provides, for example to:

* Allow blue/green apps to have the same identity.
* Use Kf with a federated identity system.
* Provide custom image pull credentials for a specific app.

You can enable this by adding the `apps.kf.dev/service-account-name`
[annotation to your app manifest]({{<relref "manifest#metadata-fields">}}).
The value should be the name of the KSA you want the application and tasks to use.

Example:

```yaml
applications:
- name: my-app
  metadata:
    annotations:
      "apps.kf.dev/service-account-name": "override-sa-name"
```

Limitations:

* Only KSAs within the same Kubernetes namespace--corresponding to a Kf Space--are allowed.
* The KSA must exist and be readable by Kf, otherwise the app will not deploy.
* The KSA or the cluster must have permission to pull the application's container images.

## Additional resources

* If you use GKE, learn how to
  [inject apps with a Google Service Account credential](https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity).
* Learn how to use [federated identity](https://cloud.google.com/iam/docs/workload-identity-federation) in GCP to allow authenticating
  KSAs to GCP infrastructure.