---
title: "Awwvision App "
linkTitle: "Awwvision App"
weight: 20
description: >
  Learn how to deploy the Awwvision reference application originally built
  for Cloud Foundry and the Google Cloud service broker.
---

[cloud-vision]: https://cloud.google.com/vision/
[aww]: https://reddit.com/r/aww
[gcs]: https://cloud.google.com/storage/
Awwvision is a Spring Boot application that uses the [Google Cloud Vision API][cloud-vision] to classify
images from Reddit's [/r/aww][aww] subreddit, store the images and classifications in [Google Cloud
Storage (GCS)][gcs], and display the labeled results in a web app. It uses the GCP Service Broker to
authenticate with the Vision and Storage APIs.

[original]: https://github.com/GoogleCloudPlatform/cloud-vision/tree/master/python/awwvision
Awwvision was inspired by the [original Python version][original].

{{% alert title="Costs Money" color="warning" %}}
Following these instructions will result in creation and use of a GCS bucket
that will incur billing charges on your GCP account.
{{% /alert %}}

## Prerequisites

[google-broker]: /docs/operators/service-brokers/google-cloud/
[install-instructions]: /docs/getting-started/install
1. **Kf**: Follow [these instructions][install-instructions] to create a new Kubernetes cluster with Kf installed.
1. **Google Service Broker**: Follow [these instructions][google-broker] to
   install.
1. **`git`**: Git is required to clone a repository.

## Stage the app

1. Clone the reference app source and go to the Awwvision app's dir:

```sh
git clone https://github.com/GoogleCloudPlatform/service-broker-samples.git
cd service-broker-samples/awwvision
```

1. Push the app, adding the `--no-start` flag to ensure it doesn't start:

	```sh
	kf push awwvision --no-start
	```

## Create and bind services

1. Create the GCS bucket service and bind it:

	```sh
    kf create-service google-storage standard awwvision-storage
    kf bind-service awwvision awwvision-storage \
      -c '{"role":"storage.objectAdmin"}'
	```

1. Create the Cloud Vision/ML service and bind it:

    ```sh
    kf create-service google-ml-apis default ml
    kf bind-service awwvision ml \
      -c '{"role":"ml.viewer"}'
    ```

1. Retrieve and note the project ID and bucket name:

    ```sh
    kf vcap-services awwvision \
      | jq '. | to_entries[] | select(.key=="google-storage") | .value[0].credentials | {project: .ProjectId, bucket: .bucket_name}'
    ```

    {{% alert title="Save this output" color="warning" %}}
    You will need the project ID and bucket name from the previous command to
    cleanup created resources once you finish the tutorial.
    {{% /alert %}}

## Deploy and test the app

1. Once the application is bound, you can start it:

	```sh
	kf start awwvision
	```

1. Proxy to the app:

	```sh
	kf proxy awwvision
	```

1. Navigate to http://localhost:8080/reddit in your browser to trigger
   classification of images from Reddit's [r/aww][aww].

1. The page will display "Scrape completed." once it is done. From there, visit
   http://localhost:8080 to view your images!

## Destroy

[cloud-console]: https://console.cloud.google.com/storage/browser
1. Navigate to the [Cloud Storage Console][cloud-console], then select the
   project ID and bucket name you noted when you previously bound the service.
   Delete all objects in the bucket.

1. Unbind and delete the storage and ML services

    ```sh
    kf unbind-service awwvision ml
    kf unbind-service awwvision awwvision-storage
    kf delete-service ml
    kf delete-service awwvision-storage
    ```

1. Delete the app:

    ```sh
    kf delete awwvision
    ```
