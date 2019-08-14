---
title: "Link Shortener App "
linkTitle: "Link Shortener App"
weight: 20
description: >
  Learn how to deploy the Link Shortener reference application originally built
  for Cloud Foundry and the Google Cloud service broker.
---

This example Spring application is an enterprise-ready link-shortener.
The URL shortener has the following behaviour:
 
 * URLs internal to your own domain get automatically redirected.
 * URLs on the Internet are scanned for vulnerabilities:
   * If the URL is malicious, the redirect is blocked.
   * If the URL is benign, the user is shown a preview of the site and clicking
through opens in a new tab to protect privacy.
   * If the vulnerability scanner is down, a warning is shown.

{{% alert title="Costs Money" color="warning" %}}
Following these instructions will result in creation of a CloudSQL database instance
that will incur billing charges on your GCP account.
{{% /alert %}}

## Prerequisites

[google-broker]: /docs/operators/service-brokers/google-cloud/
1. A Kf cluster and the `kf` CLI installed on your workstation
1. **Google Service Broker**: Follow [these instructions][google-broker] to
   install.
1. **`git`**: Git is required to clone a repository.

## Stage the app

1. Clone the reference app source and go to the Link Shortener app's dir:

```sh
git clone https://github.com/GoogleCloudPlatform/service-broker-samples.git
cd service-broker-samples/link-shortener
```

1. Open `manifest.yml` in your editor and add `BP_JAVA_VERSION: 8.*` to the
   `env` section. This will ensure Java 8 is used to build and run the app you deploy.

    The resulting `manifest.yml` should resemble:

   ```sh
	---
	applications:
	- name: link-shortener
	  memory: 1024M
	  instances: 2
	  env:
		GOOGLE_API_KEY: "YOUR_API_KEY_HERE"
		INTERNAL_DOMAIN_SUFFIX: "YOUR_DOMAIN_HERE"
		BP_JAVA_VERSION: 8.*
	```

1. Open `pom.xml`, remove the `<parent>...</parent>` element, and replace it with:

	```sh
    <parent>
      <groupId>org.springframework.boot</groupId>
      <artifactId>spring-boot-starter-parent</artifactId>
      <version>2.1.7.RELEASE</version>
    </parent>
    ```

1. Push the app, adding a parameter to ensure it doesn't start:

	```sh
	kf push link-shortener --no-start
	```

## Create and bind a database

1. Create the service:

	```sh
	kf create-service google-cloudsql-mysql mysql-db-g1-small short-links-db
	```

1. Wait for it to spin up by repeatedly running:

	```sh
	kf service short-links-db
	```

1. Bind it when ready:

	```sh
	kf bind-service link-shortener short-links-db -c '{"role":"cloudsql.editor"}'
	```

## Deploy and test the app

1. Once the application is bound, you can start it:

	```sh
	kf start link-shortener
	```

1. Proxy to the app:

	```sh
	kf proxy link-shortener
	```

1. Navigate to http://localhost:8080 in your browser to view and use the app:

## Screenshots:

![Landing Page](landing.png)

## Destroy

1. Unbind and delete the CloudSQL database:

    ```sh
    kf unbind-service link-shortener short-links-db
    kf delete-service short-links-db
    ```

1. Delete the app:

    ```sh
    kf delete link-shortener
    ```

