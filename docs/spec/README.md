# Specification

These documents aim to build a mapping between Kubernetes objects and the
objects that make up the Cloud Foundry API.

Generally, we prefer mapping CF objects onto Knative primitives use labels to
preserve metadata. We prefer a strict policy and invariants rather than creating
CRDs to reduce workload and establish a precedent with.

Additionally, we attempt to break up these objects into self-contained units
that form a DAG to prevent tight coupling and feedback loops that are difficult
to reason about.

Each entry contains a short description of the feature, the actions available
in Cloud Foundry through the Cloud Controller and the mapping into `kf`.

## Structure

* Spec documents are named after their CF object.
* Nested objects should be named using the `_` as a separator in file names and
  a `::` separator in titles. For example, `apps_environment.md` and
  `Apps::Environment`.
* Specs should form a [DAG](https://en.wikipedia.org/wiki/Directed_acyclic_graph)
  so the system can be built and deployed in layers. For example, a spec
  describing apps MUST NOT reference the SSH or Logging components.
