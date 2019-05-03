# Service Catalog

Service Catalog is a Kubernetes Incubator project that provides a
Kubernetes-native workflow for integrating with
[Open Service Brokers](https://www.openservicebrokerapi.org/)
to provision and bind to application dependencies like databases, object
storage, message-oriented middleware, and more.

For more information,
[visit the project on github](https://github.com/kubernetes-incubator/service-catalog).

## Prerequisites

- Kubernetes 1.7+ with Beta APIs enabled
- `charts/catalog` already exists in your local machine

## Installing the Chart

To install the chart with the release name `catalog`:

```bash
$ helm install . --name catalog --namespace catalog
```

## Uninstalling the Chart

To uninstall/delete the `catalog` deployment:

```bash
$ helm delete --purge catalog
```

The command removes all the Kubernetes components associated with the chart and
deletes the release.

## Configuration

The following tables lists the configurable parameters of the Service Catalog
chart and their default values.

| Parameter | Description | Default |
|-----------|-------------|---------|
| `image` | apiserver image to use | `quay.io/kubernetes-service-catalog/service-catalog:v0.1.43` |
| `imagePullPolicy` | `imagePullPolicy` for the service catalog | `Always` |
| `apiserver.replicas` | `replicas` for the service catalog apiserver pod count | `1` |
| `apiserver.updateStrategy` | `updateStrategy` for the service catalog apiserver deployments | `RollingUpdate` |
| `apiserver.minReadySeconds` | how many seconds an apiServer pod needs to be ready before killing the next, during update | `1` |
| `apiserver.annotations` | Annotations for apiserver pods | `{}` |
| `apiserver.nodeSelector` | A nodeSelector value to apply to the apiserver pods. If not specified, no nodeSelector will be applied | |
| `apiserver.aggregator.priority` | Priority of the APIService. | `100` |
| `apiserver.aggregator.groupPriorityMinimum` | The minimum priority the group should have. | `10000` |
| `apiserver.aggregator.versionPriority` | The ordering of this API inside of the group | `20` |
| `apiserver.tls.requestHeaderCA` | Base64-encoded CA used to validate request-header authentication, when receiving delegated authentication from an aggregator. If not set, the service catalog API server will inherit this CA from the `extension-apiserver-authentication` ConfigMap if available. | `nil` |
| `apiserver.service.type` | Type of service; valid values are `LoadBalancer` , `NodePort` and `ClusterIP` | `NodePort` |
| `apiserver.service.nodePort.securePort` | If service type is `NodePort`, specifies a port in allowable range (e.g. 30000 - 32767 on minikube); The TLS-enabled endpoint will be exposed here | `30443` |
| `apiserver.service.clusterIP` | If service type is ClusterIP, specify clusterIP as `None` for `headless services` OR specify your own specific IP OR leave blank to let Kubernetes assign a cluster IP |  |
| `apiserver.storage.type` | The storage backend to use; the only valid value is `etcd`, left for other storages support in future, e.g. `crd` | `etcd` |
| `apiserver.storage.etcd.useEmbedded` | If storage type is `etcd`: Whether to embed an etcd container in the apiserver pod; THIS IS INADEQUATE FOR PRODUCTION USE! | `true` |
| `apiserver.storage.etcd.servers` | If storage type is `etcd`: etcd URL(s); override this if NOT using embedded etcd. Only etcd v3 is supported. | `http://localhost:2379` |
| `apiserver.storage.etcd.image` | etcd image to use | `quay.io/coreos/etcd:latest` |
| `apiserver.storage.etcd.imagePullPolicy` | `imagePullPolicy` for etcd | `Always` |
| `apiserver.storage.etcd.persistence.enabled` | Enable persistence using PVC | `false` |
| `apiserver.storage.etcd.persistence.storageClass` | PVC Storage Class | `nil` (uses alpha storage class annotation) |
| `apiserver.storage.etcd.persistence.accessMode` | PVC Access Mode | `ReadWriteOnce` |
| `apiserver.storage.etcd.persistence.size` | PVC Storage Request | `4Gi` |
| `apiserver.storage.etcd.resources` | Resources allocation (Requests and Limits) | `{requests: {cpu: 100m, memory: 30Mi}, limits: {cpu: 100m, memory: 40Mi}}` |
| `apiserver.verbosity` | Log level; valid values are in the range 0 - 10 | `10` |
| `apiserver.auth.enabled` | Enable authentication and authorization | `true` |
| `apiserver.audit.activated` | If true, enables the use of audit features via this chart. | `false` |
| `apiserver.audit.logPath` | If specified, audit log goes to specified path. | `"/tmp/service-catalog-apiserver-audit.log"` |
| `apiserver.healthcheck.enabled` | Enable readiness and liveliness probes | `true` |
| `apiserver.serviceAccount` | Service account. | `service-catalog-apiserver` |
| `apiserver.serveOpenAPISpec` | If true, makes the API server serve the OpenAPI schema | `false` |
| `apiserver.resources` | Resources allocation (Requests and Limits) | `{requests: {cpu: 100m, memory: 20Mi}, limits: {cpu: 100m, memory: 30Mi}}` |
| `controllerManager.replicas` | `replicas` for the service catalog controllerManager pod count | `1` |
| `controllerManager.updateStrategy` | `updateStrategy` for the service catalog controllerManager deployments | `RollingUpdate` |
| `controllerManager.minReadySeconds` | how many seconds a controllerManager pod needs to be ready before killing the next, during update | `1` |
| `controllerManager.annotations` | Annotations for controllerManager pods | `{}` |
| `controllerManager.nodeSelector` | A nodeSelector value to apply to the controllerManager pods. If not specified, no nodeSelector will be applied | |
| `controllerManager.healthcheck.enabled` | Enable readiness and liveliness probes | `true` |
| `controllerManager.verbosity` | Log level; valid values are in the range 0 - 10 | `10` |
| `controllerManager.resyncInterval` | How often the controller should resync informers; duration format (`20m`, `1h`, etc) | `5m` |
| `controllerManager.brokerRelistInterval` | How often the controller should relist the catalogs of ready brokers; duration format (`20m`, `1h`, etc) | `24h` |
| `controllerManager.brokerRelistIntervalActivated` | Whether or not the controller supports a --broker-relist-interval flag. If this is set to true, brokerRelistInterval will be used as the value for that flag. | `true` |
| `controllerManager.profiling.disabled` | Disable profiling via web interface host:port/debug/pprof/ | `false` |
| `controllerManager.profiling.contentionProfiling` | Enables lock contention profiling, if profiling is enabled | `false` |
| `controllerManager.leaderElection.activated` | Whether the controller has leader election enabled | `false` |
| `controllerManager.serviceAccount` | Service account | `service-catalog-controller-manager` |
| `controllerManager.apiserverSkipVerify` | Controls whether the API server's TLS verification should be skipped | `true` |
| `controllerManager.enablePrometheusScrape` | Whether the controller will expose metrics on /metrics | `false` |
| `controllerManager.resources` | Resources allocation (Requests and Limits) | `{requests: {cpu: 100m, memory: 20Mi}, limits: {cpu: 100m, memory: 30Mi}}` |
| `useAggregator` | whether or not to set up the controller-manager to go through the main Kubernetes API server's API aggregator | `true` |
| `rbacEnable` | If true, create & use RBAC resources | `true` |
| `originatingIdentityEnabled` | Whether the OriginatingIdentity feature should be enabled | `true` |
| `asyncBindingOperationsEnabled` | Whether or not alpha support for async binding operations is enabled | `false` |
| `namespacedServiceBrokerDisabled` | Whether or not alpha support for namespace scoped brokers is disabled | `false` |

Specify each parameter using the `--set key=value[,key=value]` argument to
`helm install`.

Alternatively, a YAML file that specifies the values for the parameters can be
provided while installing the chart. For example:

```bash
$ helm install . --name catalog --namespace catalog --values values.yaml
```
