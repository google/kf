# Kf

This is not an officially supported Google product.

## Getting started the manual way

Follow the install instructions at https://cloud.google.com/migrate/kf/docs/ to create a GKE cluster,
install Kf into it, and deploy an app with the `kf` CLI.

## Deploy a local Kf install to a new cluster

If you need to set up a new development cluster run the following command:

```sh
./hack/deploy-dev-release.sh
```

It will fetch all your local sources and kick off a Cloud Build that builds
a version of Kf, creates a GKE cluster and installs the Kf version onto it.

## Iterative development

**Building the CLI:**

```sh
$ ./hack/build.sh
```

**Installing Kf server-side components:**

***With the Operator***

Kf is installed through the Operator. Operator code is in the `operator` folder.
To apply current change to the server, run `hack/apply-kf-with-operator.sh`.
This script will build `kf` with `ko`, copy the built yaml file into the operator
data folder(operator/cmd/manager/kodata/kf/), then build and apply the `operator`.


***Without the Operator***

Kf can be installed independently. To do so, you will need to disable the Operator first.
One way to do so is to run:

```sh
kubectl patch kfsystem kfsystem --type='json' -p="[{'op':'replace','path':'/spec/kf/enabled','value':false}]"
```

We use [ko](https://github.com/google/ko) for rapid development
and during the release process to build a full set of `kf` images
and installation YAML. Run the following to stage local changes on
a targeted cluster:

```sh
$ ./hack/ko-apply.sh
```

This will build any images required by `config/`, upload them to the provided
registry, and apply the resulting configuration to the current cluster.


**Verify the installation of Kf components:**

***Kf CLI***

Kf CLI can be downlowed from [official releases](https://cloud.google.com/migrate/kf/docs/2.11/downloads) or build locally. 

To build Kf CLI locally, run `hack/build.sh`. A executable `kf` will be generated under `bin`.

***Kf server side component**

Kf has a built-in self diagnostic tool called `Kf doctor`. Run `kf doctor` to run through the diagnotics to make sure the Kf server side component and dependencies are properly installed.


**Run tests:**

All tests can be run using the script `hack/test.sh`. Integration tests can be skipped by setting the environment
variable `SKIP_INTEGRATION` to `true`.

Optionally, unit tests can be run separately with script `hack/unit-test.sh`. 

Integration tests can be run with script `hack/integration-test.sh`. Integrationt tests requires a Kubernetes cluster
with Kf server component installed.