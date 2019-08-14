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
authenticate to the Vision and Storage APIs.

[original]: https://github.com/GoogleCloudPlatform/cloud-vision/tree/master/python/awwvision
Awwvision was inspired by the [original Python version][original].

{{% alert title="Costs Money" color="warning" %}}
Following these instructions will result in creation and use of a GCS bucket
that will incur billing charges on your GCP account.
{{% /alert %}}

## Prerequisites

[google-broker]: /docs/operators/service-brokers/google-cloud/
1. A Kf cluster and the `kf` CLI installed on your workstation
1. **Google Service Broker**: Follow [these instructions][google-broker] to
   install.
1. **`git`**: Git is required to clone a repository.

## Stage the app

1. Clone the reference app source and go to the Awwvision app's dir:

```sh
git clone https://github.com/GoogleCloudPlatform/service-broker-samples.git
cd service-broker-samples/awwvision
```

1. Push the app, adding a parameter to ensure it doesn't start:

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
    kf bind-service awwvision ml
      -c '{"role":"ml.viewer"}'
    ```

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
