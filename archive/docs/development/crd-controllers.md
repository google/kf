# CRD Controllers

## Theory

Kubernetes allows developers to extend the objects its APIs process and store
using Custom Resource Definitions (CRDs). If you think about Kubernetes like a
database, CRDs are table definitions, and controllers are like triggers.

A controller is a process that runs in a pod and watches for changes. The
controller needs the ability to watch for changes and write back to the
Kubernetes API. When it starts, the controller tells Kubernetes what events
it's interested in. Kubernetes will tell the controller the name of objects that
have changed.

Internally, the controller should queue up these objects. For each object it
will run a reconciliation loop that looks something like the following pseudocode:

``` python
def reconcile(object_name):
    # First, get the contents of the object
    try:
      obj = kubernetes.get_object_by_name(object_name)
    except DoesNotExist:
      # the update was a delete, or the object was deleted after the update
      # entered the queue, ignore.
      return

    if obj.is_deleting():
      return # ignore

    # copy the object so we can tell if it needs an update or not
    new_obj = reconcile_object(copy(obj))

    # if the reconciliation changed the system, update it
    if new_obj.status != obj.status:
      kubernetes.update_object_status(new_obj)
```

To reconcile the fully populated object each sub-resource is reconciled
similarly to how the parent was:

``` python
def reconcile_object(obj):
  obj = apply_defaults(obj)
  obj.status.initialize_conditions() # set all unset conditions to unknown

  # upgrade the object from a potentially earlier version e.g. v1alpha1 to v1beta1
  obj = convert_up(obj)

  # for each sub_resource on obj:
  try:
    srobj = kubernetes.get_object(sub_resource)
    if obj not in srobj.owners:
      obj.status.conditions += "Error: sub-resource not owned by obj"
      continue

    reconcile_sub_resource(srobj)
  except DoesNotExist:
    # create the sub-resource based on properties from the object
    srobj = init_subresource(obj)
    kubernetes.create_object(srobj)

  # update status in parent object
  obj.status.XYZ = srobj.XYZ
  # end for
```

Besides being triggered when object changes are made, the reconciliation system
may run periodically to ensure no events were missed.

## Implementation

We will follow Knative's lead on building controllers so we can take advantage
of the work they've done generalizing them.

The main controller will be installed into the `kf` namespace by a Kubernetes
deployment in `config/controller.yaml` that references `cmd/controller` as the
container it will run.

`cmd/controller` will be the entrypoint of the controller application. It will
configure logging, initialize watchers, and initialize reconcilers for all CRDs.
For example, [Knative Serving's controller](https://github.com/knative/serving/blob/master/cmd/controller/main.go).

`pkg/reconciler/CRDNAME` will contain the main reconciliation loop for the CRD.
For example, [Knative Serving's service reconciler](https://github.com/knative/serving/tree/master/pkg/reconciler/service).
