# Operand

## Description

An Operand is a resource which is responsible for maintaining a changing
set of resources on the cluster. You can think of it as the reification
of the Manifest concept used by human Operators into a CRD.

## Spec

An Operand is configured much in the way a human configures a k8s cluster, by
providing a set of YAML. In this case, the YAML is created by another
reconciliation loop (the CloudRun Operator reconciler), and actuated by the
Operand reconciler.

Today resources are divided into two sets: the `SteadyState` and `PostInstall`
fields which contain a set of unstructured.Unstructured resources.

In addition `CheckDeploymentHealth` can be used to control whether the
controller deployment will be checked for Readiness before this Operand goes
ready.

## Status

* LatestReadyActiveOperand: Tracks latest CAO that has gone ready (this is our
  baseline 'active version').
* LatestCreatedActiveOperand: Tracks latest created CAO (this is our 'pending
  operation')
* InstalledSteadyStateGeneration: Tracks the generation of the currently applied
  SteadyState after successful install. Clear when a new SteadyState is applied
  to indicate that the currently applied SteadyState is not yet ready.
* Status bit LatestActiveOperandReady: equal to LatestCreatedActiveOperand ==
  LatestReadyActiveOperand && LatestReadyActiveOperand == desired.
* Status bit OperandInstalled: whether the manifest used by the Operand is
  applied fully to the cluster. Note that annotations are used to control the
  merge behavior.

## Reconciliation Process

### Reconcile

1. Calculate manifest
1. Compute LiveRefs (and CAO name, which is a hash of these)
1. If needed, update the LatestCreatedActiveOperand, create CAO, and reset
   LatestActiveOperandReady bit. If the previous latest created CAO was not
   ready delete it before updating LatestCreatedActiveOperand (Invariants: at
   most 2 CAOs exist per Operand at any given time, and whenever a CAO exists
   its name is stored in LatestCreatedActiveOperand or
   LatestReadyActiveOperand. This implies that LatestCreatedActiveOperand must
   be updated before CAO creation and that CAO deletion must occur before a CAO
   can be cleared from the Operand's Status).
1. Check InstalledSteadyStateGeneration to determine if a SteadyState install is
   needed. If InstalledSteadyStateGeneration does match the current Operand
   generation, perform an install of SteadyState resources.
1. If InstalledSteadyStateGeneration matches the current generation, perform the
   post-install. This is done by merging PostInstall resources with SteadyState
   resources such that PostInstall resources supercede SteadyState resources in
   the case of a conflict. This merged resource set is then applied.
1. Set the OperandInstalled bit based on the success of the operand install and
   post-install.
1. Check CAO, if ready, delete previous LatestReadyActiveOperand, and set
   LatestActiveOperandReady to true.

When a CAO is deleted, it is removed as owner from all objects it owned, and if
these objects are now orphaned (equivalently, if they are not referenced by the
latest CAO) they are deleted (immediately, or later, depending on the prune
behavior chosen by the client).

### Finalize

1. Delete all CAOs associated with this Operand.
