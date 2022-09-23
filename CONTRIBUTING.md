# How to Contribute

Thanks for considering contributing to this project!
There are just a few small guidelines you need to follow.

## Scope

Kf is a tool to migrate Cloud Foundry workloads to Kuberntes by
preserving te developer experience while updating the platform operator
experience to a modern platform.

At this point in time, we consider Kf to be largely feature complete.

### Bug fixes, patches, and documentation updates

We're willing to accept bug fixes, patches, and documentation updates
provided they're complete, have tests (where applicable), and are easy
to verify.

### New features/platforms

Kf handles many of the complexities of both Cloud Foundry and Kubernetes.
The bar for adding new features is proportional to the risk of breaking
something.

Before working on new features, reach out to the team to make sure they fit with Kf.

Here's a non-comprehensive list of considerations:

* Kf targets the most used parts of Cloud Foundry.
* Platform operators like Kubernetes, let them use `kubectl`.
* Kf is a pathway to Kubernetes -- it shouldn't provide CFisms that would be impossible to achieve outside of Kf.
* Kf targets GKE and Anthos.

## Contributor License Agreement

Contributions to this project must be accompanied by a Contributor License
Agreement (CLA). You (or your employer) retain the copyright to your
contribution; this simply gives us permission to use and redistribute your
contributions as part of the project. Head over to
<https://cla.developers.google.com/> to see your current agreements on file or
to sign a new one.

You generally only need to submit a CLA once, so if you've already submitted one
(even if it was for a different project), you probably don't need to do it
again.

## Code Reviews

All submissions, including submissions by project members, require review. We
use GitHub pull requests for this purpose. Consult
[GitHub Help](https://help.github.com/articles/about-pull-requests/) for more
information on using pull requests.

## Community Guidelines

This project follows
[Google's Open Source Community Guidelines](https://opensource.google/conduct/).
