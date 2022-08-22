---
title: Deploy Spring Music
description: >
    Learn to deploy an app with backing services.
---

These instructions will walk you through deploying the [Cloud Foundry Spring
Music](https://github.com/cloudfoundry-samples/spring-music) reference App using the Kf Cloud Service Broker.

1. **Building Java Apps from source**: The Spring Music source will be built on
   the cluster, not locally.

1. **Service broker integration**: You will create a database using the Kf Cloud Service Broker and bind the Spring Music App to it.

1. **Spring Cloud Connectors**: [Spring Cloud
   Connectors](https://cloud.spring.io/spring-cloud-connectors/) are used by the Spring Music App to
   detect things like bound CF services. They work seamlessly with Kf.

1. **Configuring the Java version**: You will specify the version of Java you
   want the buildpack to use.

## Prerequisites

[Install and configure the Kf Cloud Service Broker]({{< relref "deploying-cloud-sb" >}}).

## Deploy Spring Music

### Clone source
1. Clone the [Spring Music repo](https://github.com/cloudfoundry-samples/spring-music).

    ```sh
    git clone https://github.com/cloudfoundry-samples/spring-music.git spring-music

    cd spring-music
    ```

1. Edit `manifest.yml`, and replace `path: build/libs/spring-music-1.0.jar` with `stack: org.cloudfoundry.stacks.cflinuxfs3`. This instructs Kf to build from source using [cloud native buildpacks](https://cloud.google.com/blog/products/containers-kubernetes/google-cloud-now-supports-buildpacks) so you don't have to compile locally.

    ```sh
    ---
    applications:
    - name: spring-music
      memory: 1G
      random-route: true
      stack: org.cloudfoundry.stacks.cflinuxfs3
      env:
        JBP_CONFIG_SPRING_AUTO_RECONFIGURATION: '{enabled: false}'
    #    JBP_CONFIG_OPEN_JDK_JRE: '{ jre: { version: 11.+ } }'
    ```

### Push Spring Music with no bindings

1. Create and target a Space.

    ```sh
    kf create-space test

    kf target -s test
    ```

1. Deploy Spring Music.

    ```sh
    kf push spring-music
    ```

1. Use the proxy feature to access the deployed App.

  1. Start the proxy:

      ```sh
      kf proxy spring-music
      ```

  1. Open `http://localhost:8080` in your browser:

      {{< figure src="./sm1.png" alt="Screenshot of Spring Music showing no profile." >}}

      The deployed App includes a UI element showing which (if any) Spring profile is being used. No profile is being used here, indicating an in-memory database is in use.

### Create and bind a database

1. Create a PostgresSQL database from the marketplace.

    {{% note %}}You must set the `COMPUTE_REGION` and `VPC_NAME` variables so Kf Cloud Service Broker knows where to provision your instance, and authorize the VPC Kf Apps run on to access it.{{% /note %}}

    ```sh
    kf create-service csb-google-postgres small spring-music-postgres-db -c '{"region":"COMPUTE_REGION","authorized_network":"VPC_NAME"}'
    ```

1. Bind the Service with the App.

    ```sh
    kf bind-service spring-music spring-music-postgres-db
    ```

1. Restart the App to make the service binding available via the VCAP_SERVICES environment variable.

    ```sh
    kf restart spring-music
    ```

1. (Optional) View the binding details.

    ```sh
    kf bindings
    ```

1. Verify the App is using the new binding.

    1. Start the proxy:

        ```sh
        kf proxy spring-music
        ```

    1. Open `http://localhost:8080` in your browser:

        {{< figure src="./sm2.png" alt="Screenshot of Spring Music showing a profile." >}}

        You now see the Postgres profile is being used, and we see the name of our Service we bound the App to.

## Clean up

1. Unbind and delete the PostgreSQL service:

    ```sh
    kf unbind-service spring-music spring-music-db

    kf delete-service spring-music-db
    ```

1. Delete the App:

    ```sh
    kf delete spring-music
    ```
