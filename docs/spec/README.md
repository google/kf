# Specification

These documents aim to build a mapping from Kubernetes objects and workflows
to their equivalent objects and workflows that make up the Cloud Foundry API.

CF objects should be mapped to well-known Knative primitives when possible,
with labels preserving metadata across the components, as opposed to building
new CRDs. This strict usage policy allows kf to be as thin a layer on top of
Knative as possible.

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
