---
title: "Quickstart"
weight: 10
---

In this quickstart, you will deploy a sample Cloud Foundry app on an existing Kf cluster.

## Push an application

### Prerequisites

The following are required to complete this section:

1. The `kf` CLI installed and in your path.
1. You have connected to the Kf Kubernetes cluster:

    ```sh
    gcloud container clusters get-credentials CLUSTER_NAME \
        --project=CLUSTER_PROJECT_ID \
        --zone=CLUSTER_LOCATION
    ```

1. The `git` CLI installed and in your path.

### Prepare space

1. Create new space:

    ```sh
    kf create-space test-space
    ```

1. Target the space:

    ```sh
    kf target -s test-space
    ```

### Push the Cloud Foundry test app


1. Clone the [test-app repo](https://github.com/cloudfoundry-samples/test-app).

    ```sh
    git clone https://github.com/cloudfoundry-samples/test-app go-test-app

    cd go-test-app
    ```

1. Push the app.

    {{< note >}}It will take a few minutes for the build to complete.{{< /note >}}

    ```sh
    kf push test-app
    ```

1. Get the application's URL.

    ```sh
    kf apps
    ```

    {{< note >}}The app will have a random route by default.{{< /note >}}


1. Open the URL in your browser where you should see the app running.

    {{< figure src="./app-quickstart-success.png" title="Successful app push on Kf" >}}

## Clean up

These steps should return the cluster to the starting state.

1. Delete the application.

    ```sh
    kf delete test-app
    ```

1. Delete the Space.

    ```sh
    kf delete-space test-space
    ```
