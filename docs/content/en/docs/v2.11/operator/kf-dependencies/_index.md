---
title: Kf dependencies and architecture
---

Kf requires Kubernetes and several other OSS projects to
run. Some of the dependencies are satisfied with Google-managed services&mdash;for
example, GKE provides Kubernetes.

## Dependencies {#dependencies}

* [GKE](https://cloud.google.com/kubernetes-engine/docs/release-notes)
* [Anthos Service Mesh](https://cloud.google.com/service-mesh/docs)
* [Tekton Pipelines](https://github.com/tektoncd/pipeline)

{{<figure src="./kf-components-diagram.svg" alt="Diagram showing how Kf components interact.">}}


## Get CRD details

Kf supports the `kubectl`  subcommand `explain`. It allows you to list the fields in Kf CRDs to understand how to create Kf objects via automation instead of manually via the CLI. This command is designed to be used with [ConfigSync](https://github.com/GoogleContainerTools/kpt-config-sync) to automate creation and management of resources like Spaces across many clusters. You can use this against any of the [component `kinds` below](#components).

In this example, we examine the `kind` called `space` in the `spaces` CRD:

```sh
kubectl explain space.spec
```

The output looks similar to this:

```
$ kubectl explain space.spec
KIND:     Space
VERSION:  kf.dev/v1alpha1

RESOURCE: spec <Object>

DESCRIPTION:
     SpaceSpec contains the specification for a space.

FIELDS:
   buildConfig  <Object>
     BuildConfig contains config for the build pipelines.

   networkConfig        <Object>
     NetworkConfig contains settings for the space's networking environment.

   runtimeConfig        <Object>
     RuntimeConfig contains settings for the app runtime environment.
```

## Kf components {#components}

Kf installs several of its own Kubernetes
[custom resources](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/)
and [controllers](https://kubernetes.io/docs/concepts/workloads/controllers/).
The custom resources effectively serve as the Kf API and
are used by the `kf` CLI to interact with the system. The controllers use
Kf's CRDs to orchestrate the other components in the
system.

You can view the CRDs installed and used by Kf by running
the following command:

```sh
kubectl api-resources --api-group=kf.dev
```

The output of that command is as follows:

```none
NAME                      SHORTNAMES   APIGROUP   NAMESPACED   KIND
apps                                   kf.dev     true         App
builds                                 kf.dev     true         Build
clusterservicebrokers                  kf.dev     false        ClusterServiceBroker
routes                                 kf.dev     true         Route
servicebrokers                         kf.dev     true         ServiceBroker
serviceinstancebindings                kf.dev     true         ServiceInstanceBinding
serviceinstances                       kf.dev     true         ServiceInstance
spaces                                 kf.dev     false        Space
```

### Apps {#apps}

Apps represent a [twelve-factor application](https://12factor.net/)
deployed to Kubernetes. They encompass source code, configuration, and the
current state of the application. Apps are responsible for reconciling:

* Kf Builds
* Kf Routes
* Kubernetes Deployments
* Kubernetes Services
* Kubernetes ServiceAccounts
* Kubernetes Secrets

You can list Apps using Kf or `kubectl`:

```sh
kf apps

kubectl get apps -n space-name
```

### Builds {#builds}

Builds combine the source code and build configuration for Apps. They provision
Tekton TaskRuns with the correct steps to actuate a Buildpack V2, Buildpack V3,
or Dockerfile build.

You can list Builds using Kf or `kubectl`:

```sh
kf builds

kubectl get builds -n space-name
```

### ClusterServiceBrokers {#clusterservicebrokers}

ClusterServiceBrokers hold the connection information necessary to extend
Kf with a service broker. They are responsible for
fetching the catalog of services the broker provides and displaying them in
the output of `kf marketplace`.

You can list ClusterServiceBrokers using `kubectl`:

```sh
kubectl get clusterservicebrokers
```

### Routes {#routes}

Routes are a high level structure that contain HTTP routing rules. They are
responsible for reconciling Istio VirtualServices.

You can list Routes using Kf or `kubectl`:

```sh
kf routes

kubectl get routes -n space-name
```

### ServiceBrokers {#servicebrokers}

ServiceBrokers hold the connection information necessary to extend
Kf with a service broker. They are responsible for
fetching the catalog of services the broker provides and displaying them in
the output of `kf marketplace`.

You can list ServiceBrokers using `kubectl`:

```sh
kubectl get servicebrokers -n space-name
```

### ServiceInstanceBinding {#serviceinstancebinding}

ServiceInstanceBindings hold the parameters to create a binding on a service
broker and the credentials the broker returns for the binding. They are
responsible for calling the bind API on the broker to bind the service.

You can list ServiceInstanceBindings using Kf or `kubectl`:

```sh
kf bindings

kubectl get serviceinstancebindings -n space-name
```

### ServiceInstance {#serviceinstance}

ServiceInstances hold the parameters to create a service on a service broker.
They are responsible for calling the provision API on the broker to create the
service.

You can list ServiceInstances using Kf or `kubectl`:

```sh
kf services

kubectl get serviceinstances -n space-name
```

### Spaces {#spaces}

Spaces hold configuration information similar to Cloud Foundry organizations and
spaces. They are responsible for:

* Creating the Kubernetes Namespace that other Kf resources are provisioned into.
* Creating Kubernetes NetworkPolicies to enforce network connection policies.
* Holding configuration and policy for Builds, Apps, and Routes.

You can list Spaces using Kf or `kubectl`:

```sh
kf spaces

kubectl get spaces
```

## Kf RBAC / Permissions

The following sections list permissions for Kf and its
components to have correct access at the cluster level.
These permissions are required and enabled by default in Kf; do
not attempt to disable them.

<table>
  <thead>
  <tr>
  <th><strong>Components</strong></th>
  <th><strong>Namespace</strong></th>
  <th><strong>Service Account</strong></th>
  </tr>
  </thead>
  <tbody>
  <tr>
  <td><code>controller</code></td>
  <td>kf</td>
  <td>controller</td>
  <tr>
  <td><code>subresource-apiserver</code></td>
  <td>kf</td>
  <td>controller</td>
  </tr>
  <tr>
  <td><code>webhook</code></td>
  <td>kf</td>
  <td>controller</td>
  </tr>
  <tr>
  <td><code>appdevexperience-operator</code></td>
  <td>appdevexperience</td>
  <td>appdevexperience-operator</td>
  </tr>
  </tbody>
</table>

Note that the `appdevexperience-operator` service account has the same set of
permissions as `controller`. The operator is what deploys all Kf
components, including custom resource definitions and controllers.

### RBAC for Kf service accounts

The following `apiGroup` definitions detail which access control
permissions components in {{product_name}} have on which API groups and resources for both the `controller`
and `appdevexperience-operator` service accounts.

```yaml
- apiGroups:
  - "authentication.k8s.io"
  resources:
  - tokenreviews
  verbs:
  - create
- apiGroups:
  - "authorization.k8s.io"
  resources:
  - subjectaccessreviews
  verbs:
  - create
- apiGroups:
  - ""
  resources:
  - pods
  - services
  - persistentvolumeclaims
  - persistentvolumes
  - endpoints
  - events
  - configmaps
  - secrets
  verbs: *
- apiGroups:
  - ""
  resources:
  - services
  - services/status
  verbs:
  - create
  - delete
  - get
  - list
  - watch
- apiGroups:
  - "apps"
  resources:
  - deployments
  - daemonsets
  - replicasets
  - statefulsets
  verbs: *
- apiGroups:
  - "apps"
  resources:
  - deployments/finalizers
  verbs:
  - get
  - list
  - create
  - update
  - delete
  - patch
  - watch
- apiGroups:
  - "rbac.authorization.k8s.io"
  resources:
  - clusterroles
  - roles
  - clusterrolebindings
  - rolebindings
  verbs:
  - create
  - delete
  - update
  - patch
  - escalate
  - get
  - list
  - deletecollection
  - bind
- apiGroups:
  - "apiregistration.k8s.io"
  resources:
  - apiservices
  verbs:
  - update
  - patch
  - create
  - delete
  - get
  - list
- apiGroups:
  - "pubsub.cloud.google.com"
  resources:
  - topics 
  - topics/status
  verbs: *
- apiGroups:
  - ""
  resources:
  - namespaces
  - namespaces/finalizers
  - serviceaccounts
  verbs: 
  - get
  - list
  - create
  - update
  - watch
  - delete
  - patch
  - watch
- apiGroups:
  - "autoscaling"
  resources:
  - horizontalpodautoscalers
  verbs: 
  - create
  - delete
  - get
  - list
  - update
  - patch
  - watch
- apiGroups:
  - "coordination.k8s.io"
  resources:
  - leases
  verbs: *
- apiGroups:
  - "batch"
  resources:
  - jobs
  - cronjobs
  verbs: 
  - get
  - list
  - create
  - update
  - patch
  - delete
  - deletecollection
  - watch
- apiGroups:
  - "messaging.cloud.google.com"
  resources:
  - channels
  verbs: 
  - delete
- apiGroups:
  - "pubsub.cloud.google.com"
  resources:
  - pullsubscriptions
  verbs: 
  - delete
  - get
  - list
  - watch
  - create
  - update
  - patch
- apiGroups:
  - "pubsub.cloud.google.com"
  resources:
  - [pullsubscriptions/status
  verbs: 
  - get
  - update
  - patch
- apiGroups:
  - "events.cloud.google.com"
  resources: *
  verbs: *
- apiGroups:
  - "keda.k8s.io"
  resources: *
  verbs: *
- apiGroups:
  - "admissionregistration.k8s.io"
  resources:
  - mutatingwebhookconfigurations
  - validatingwebhookconfigurations
  verbs:
  - get
  - list
  - create
  - update
  - patch
  - delete
  - watch
- apiGroups:
  - "extensions"
  resources:
  - ingresses
  - ingresses/status
  verbs: *
- apiGroups:
  - ""
  resources: 
  - endpoints/restricted
  verbs:
  - create
- apiGroups:
  - "certificates.k8s.io"
  resources: 
  - certificatesigningrequests
  - certificatesigningrequests/approval
  - certificatesigningrequests/status
  verbs: 
  - update
  - create
  - get
  - delete
- apiGroups:
  - "apiextensions.k8s.io"
  resources:
  - customresourcedefinitions
  verbs:   
  - get
  - list
  - create
  - update
  - patch
  - delete
  - watch
- apiGroups:
  - "networking.k8s.io"
  resources: 
  - networkpolicies
  verbs: 
  - get
  - list
  - create
  - update
  - patch
  - delete
  - deletecollection
  - watch
- apiGroups:
  - ""
  resources: 
  - nodes
  verbs: 
  - get
  - list
  - watch
  - update
  - patch
- apiGroups:
  - ""
  resources: 
  - nodes/status
  verbs: 
  - patch
```

The following table lists how the RBAC permissions are used in Kf,
where:

* **view** includes the verbs: get, list, watch
* **modify** includes the verbs: create, update, delete, patch

<table>
  <thead>
  <tr>
  <th><strong>Permissions</strong></th>
  <th><strong>Reasons</strong></th>
  </tr>
  </thead>
  <tbody>
  <tr>
  <td>Can view all <code>secrets</code></td>
  <td>Kf reconcilers need to read secrets for functionalities such as space creation and service instance binding.</td>
  <tr>
  <td>Can modify <code>pods</code></td>
  <td>Kf reconcilers need to modify pods for functionalities such as building/pushing Apps and Tasks.</td>
  </tr>
  <tr>
  <td>Can modify <code>secrets</code></td>
  <td>Kf reconcilers need to modify secrets for functionalities such as building/pushing Apps and Tasks and service instance binding.</td>
  </tr>
  <tr>
  <td>Can modify <code>configmaps</code></td>
  <td>Kf reconcilers need to modify configmaps for functionalities such as building/pushing Apps and Tasks.</td>
  </tr>
  <tr>
  <td>Can modify <code>endpoints</code></td>
  <td>Kf reconcilers need to modify endpoints for functionalities such as building/pushing Apps and route binding.</td>
  </tr>
  <tr>
  <td>Can modify <code>services</code></td>
  <td>Kf reconcilers need to modify pods for functionalities such as building/pushing Apps and route binding.</td>
  </tr>
  <tr>
  <td>Can modify <code>events</code></td>
  <td>Kf controller creates and emits events for the resources
    managed by Kf.</td>
  </tr>
  <tr>
  <td>Can modify <code>serviceaccounts</code></td>
  <td>Kf needs to modify service accounts for App deployments.</td>
  </tr>
  <tr>
  <td>Can modify <code>endpoints/restricted</code></td>
  <td>Kf needs to modify endpoints for App deployments.</td>
  </tr>
  <tr>
  <td>Can modify <code>deployments</code></td>
  <td>Kf needs to modify deployments for functionalities such as pushing Apps.</td>
  </tr>
  <tr>
  <td>Can modify <code>mutatingwebhookconfiguration</code></td>
  <td>Mutatingwebhookconfiguration is needed by {{mesh_name}}, a Kf dependency, for admission webhooks.</td>
  </tr>
  <tr>
  <td>Can modify
    <code>customresourcedefinitions customresourcedefinitions/status</code></td>
  <td>Kf manages resources through Custom Resources such as Apps, Spaces and Builds.</td>
  </tr>
  <tr>
  <td>Can modify <code>horizontalpodautoscalers</code></td>
  <td>Kf supports autoscaling based on Horizontal Pod Autoscalers.</td>
  </tr>
  <tr>
  <td>Can modify <code>namespace/finalizer</code></td>
  <td>Kf needs to set owner reference of webhooks.</td>
  </tr>
  </tbody>
</table>

## Third-party libraries {#third_party_libraries}

Third-party library source code and licenses can be found in the `/third_party`
directory of any Kf container image.

You can also run `kf third-party-licenses` to view the third-party licenses for
the version of the Kf CLI that you downloaded.
