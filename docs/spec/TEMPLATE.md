# ConceptName

<!--
Enter an English description of the concept here and a rough explanation of how
it's mapped to the Kubernetes model.
-->

Cloud Foundry ConceptName are mapped to Kubernetes BackingObjects.

<!--
Populate this table with the action available on the Cloud controller in the
Action column and the Kubernetes verb and noun in the right hand column.

If an action is not supported in kf, denote it with _not_mapped_.

This table is useful for designing RBAC.
-->
| Action | `kf` Action |
|--------|-------------|
| Create | create Foo |
| Get | get Foo |
| List | list Foo |
| Set X | get Foo; update `path.to.X`; update Foo |

## Metadata

This section contains meta data mappings for the Kubernetes object.

### Labels

<!--
This list is filled with the valid labels kf supports on the object.
-->

* `app.kubernetes.io/managed-by` - set to be `kf`.

### Annotations

<!--
This list is filled with the annotations kf supports on the object.
-->

* _None_

### Parent

<!--
Make notes here about what the ownership of this object should be to ensure a
clean lifecycle.
-->

This resource has namespace as an implicit parent.

## Required Policies

<!--
Make notes here about policies (webhooks) that should be run when an object is
created/updated to ensure it remains valid.
-->

These policies must be enforced to ensure the model remains valid.

* _No required policies._

## Optional Policies

<!--
Make notes here about policies (webhooks) that could be enabled/added in certain
situations to enhance customer experience.

For example, there could be a policy that only docker images from certain
repositories can be allowed to run.
-->

* _No optional policies._
