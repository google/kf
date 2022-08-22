---
title: Deploy Spring Cloud Config
description: >
  Learn how to deploy Spring Cloud Config as a configuration server.
---

This document shows how to deploy Spring Cloud Config in a Kf cluster.

[Spring Cloud Config](https://cloud.spring.io/spring-cloud-config/reference/html/)
provides a way to decouple application code from its runtime configuration.
The Spring Cloud Config configuration server can read configuration files from Git
repositories, the local filesystem, [HashiCorp Vault servers](https://www.vaultproject.io/),
or [Cloud Foundry CredHub](https://docs.cloudfoundry.org/credhub/).
Once the configuration server has read the configuration, it can format and serve
that configuration as YAML, [Java Properties](https://docs.oracle.com/cd/E23095_01/Platform.93/ATGProgGuide/html/s0204propertiesfileformat01.html),
or JSON over HTTP.

## Before you begin

You will need a cluster with Kf installed and access to the Kf CLI.

Additionally, you will need the following software:

- **`git`**: Git is required to clone a repository.

## Download the Spring Cloud Config configuration server

To download the configuration server source:

1.  Open a terminal.
1.  Clone the source for the configuration server:

    ```sh
    git clone --depth 1 "https://github.com/google/kf"
    ```

## Configure and deploy a configuration server

To update the settings for the instance:

1. Change directory to `spring-cloud-config-server`:

    ```sh
    cd kf/spring-cloud-config-server
    ```
1.  Open `manifest.yaml`.
1.  Change the `GIT_URI` environment variable to the URI of your Git configuration server.
1.  Optionally, change the name of the application in the manifest.
1.  Optionally, [configure additional properties or alternative property sources](https://cloud.spring.io/spring-cloud-config/reference/html/#_environment_repository)
    by editing `src/main/resources/application.properties`.
1.  Deploy the configuration server without an external route. If you changed
    the name of the application in the manifest, update it here:

    ```sh
    kf push --no-route spring-cloud-config
    ```

{{< note >}} The default configuration is not production ready. The `README.md` file contains
additional steps to take if you want to deploy the application to production.{{< /note >}}

## Bind applications to the configuration server

You can create a [user provided service]({{< relref "user-provided-services.md" >}})
to bind the deployed configuration server to other Kf
applications in the same cluster or namespace.

How you configure them will depend on the library you use:


*  Applications using Pivotal's Spring Cloud services client library

  Existing PCF applications that use [Pivotal's Spring Cloud Services client library](https://github.com/pivotal-cf/spring-cloud-services-starters)
  can be bound using the following method:

  1.  Create a user provided service named <var>config-server</var>. This step
      only has to be done once per configuration server:

      ```sh
      kf cups config-server -p '{"uri":"http://spring-cloud-config"}' -t configuration
      ```

      {{< note >}}If you want to use a configuration server in a different space,
      change the URI [to include the space]({{< relref "service-discovery#how_to_use_service_discovery_with" >}}).{{< /note >}}

  2.  For each application that needs get credentials, run:

      ```sh
      kf bind-service application-name config-server

      kf restart application-name
      ```

      This will create an entry into the `VCAP_SERVICES` environment variable for
      the configuration server.

*  Other applications

  Applications that can connect directly to a Spring Cloud Config configuration
  server should be configured to access it using its cluster internal URI:

  ```none
  http://spring-cloud-config
  ```

  {{< note >}}If you want to use a configuration server in a different space, change the URI
  [to include the space name]({{< relref "service-discovery#how_to_use_service_discovery_with" >}}).{{< /note >}}

  *  For Spring applications that use the [Spring Cloud Config client library](https://cloud.spring.io/spring-cloud-config/multi/multi__spring_cloud_config_client.html)
     can set the `spring.cloud.config.uri` property in the appropriate location
     for your application. This is usually an `application.properties` or
     `application.yaml` file.
  *  For other frameworks, see your library's reference information.

## Delete the configuration server

To remove a configuration server:

1. Remove all bindings to the configuration server running the following commands for each bound application:

    ```sh
    kf unbind-service application-name config-server

    kf restart application-name
    ```

2. Remove the service entry for the configuration server:

    ```sh
    kf delete-service config-server
    ```

3. Delete the configuration server application:

    ```sh
    kf delete spring-cloud-config
    ```

## What's next

+   Read more about the [types of configuration sources](https://cloud.spring.io/spring-cloud-config/reference/html/#_environment_repository)
    Spring Cloud Config supports.
+   Learn about [the structure of the `VCAP_SERVICES` environment variable]({{< relref "app-runtime#vcapservices" >}})
    to understand how it can be used for service discovery.