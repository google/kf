# ActiveOperand

## Description

An ActiveOperand (or AO, pronounced ow) is a resource which is responsible
for owning a number of namespaced objects. It can generally be thought of
as a shared lifetime for a set of objects within its namespace. An object
is considered live if it is referenced by at least one ActiveOperand
(or at the cluster scope, by a [ClusterActiveOperand](../clusteractiveoperand/index.md)).

## Spec

The spec of an AO is relatively simple, containing a list of LiveRefs (which
point to particular objects via GVR and namespace/name).

## Status

A single bit is used to control whether this AO has completed injecting itself
as owner. This is the `OwnerRefsInjected` status.

## Reconciliation Process

### Reconcile

Attempt to add self into OwnerReference list of all objects provided.

Set `OwnerRefsInjected` based on success of these changes.


### Finalize

None. Kubernetes garbage collection is used to ensure children are deleted.
