# ClusterActiveOperand

## Description

An ClusterActiveOperand (or CAO, pronounced cow) is a resource which is responsible
for owning a number of arbitrary objects. It can generally be thought of
as a shared lifetime for a set of objects within a single cluster. An object
is considered live if it is referenced by at least one ClusterActiveOperand
(or at the namespace scope, by a [ActiveOperand](../activeoperand/index.md)).

A ClusterActiveOperand creates and owns one ActiveOperand per namespace it
references.

## Spec

The spec of an CAO is relatively simple, containing a list of LiveRefs (which
point to particular objects via GVR and namespace/name).

## Status

There are 2 status bits:

* OwnerRefsInjected: set based on whether all cluster scoped objects have had
  this CAO injected as an owner.
* NamespaceDelegatesReady: set based on whether all child AOs are Ready.

A CAO is only Ready when both of these are true.

In addition, the Status is used to track all child AOs, and contains the
cluster scoped LiveRefs set in the Spec.

## Reconciliation Process

### Reconcile

Create one child AO per namespace.

Set `NamespaceDelegatesReady` based on ready state of children.

Attempt to add self into OwnerReference list of all objects provided.

Set `OwnerRefsInjected` based on success of these changes.


### Finalize

Because garbage collection does not work across namespaces (or between cluster
and namespace scoped objects) a custom finalizer 'adapts' the CAO to AO deletion
process. During this phase, the CAO simply attempts to repeatedly delete all AOs
it references.
