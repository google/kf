# Space

<!--
Enter an English description of the concept here and a rough explanation of how
it's mapped to the Kubernetes model.
-->

Spaces in CF are mapped to Kubernetes Namespaces. The `kf` namespace is reserved.

`kf` will only treat Namespaces with the label `app.kubernetes.io/managed-by=kf`
as Spaces for the purposes of executing webhooks and listing.

<!--
Populate this table with the action available on the Cloud controller in the
Action column and the Kubernetes verb and noun in the right hand column.

This table is useful for designing RBAC.
-->
| Action | `kf` Action |
|--------|-------------|
| Create | create namespace |
| Get | get namespace |
| List | list namespace |
| Update a space | update namespace |
| Get assigned isolation segment | _Not mapped_ |
| Manage isolation segment | _Not mapped_ |

## Metadata

This section contains metadata mappings for the Kubernetes object.

### Labels

<!--
This list is filled with the valid labels kf supports on the object.
-->

* `app.kubernetes.io/managed-by` - set to be `kf`.

### Annotations

<!--
This list is filled with the annotations kf supports on the object.
-->

### Parent

<!--
Make notes here about what the ownership of this object should be to ensure a
clean lifecycle.
-->

This resource has no parent.

## Required Policies

<!--
Make notes here about policies (webhooks) that should be run when an object is
created/updated to ensure it remains valid.
-->

* _No required policies._

## Optional Policies

<!--
Make notes here about policies (webhooks) that could be enabled/added in certain
situations to enhance customer experience.

For example, there could be a policy that only docker images from certain
repositories can be allowed to run.
-->

* _No optional policies._
