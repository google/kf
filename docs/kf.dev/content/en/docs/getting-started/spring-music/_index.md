---
title: "Deploy Spring Music"
linkTitle: "Deploy Spring Music"
weight: 20
description: >
  Learn how to deploy the Spring Music Cloud Foundry reference application
  with Kf.
---

[spring-music]: https://github.com/cloudfoundry-samples/spring-music
These instructions will walk you through deploying the [Cloud Foundry Spring Music][spring-music]
reference app with Kf, demonstrating a few things along the way:

1. **Building Java apps from source**: The Spring Music source will be built on
   the cluster, not locally.

1. **Service broker integration**: You will create and bind a Postgresql
   database to the Spring Music app.

[spring-cloud-connectors]: https://cloud.spring.io/spring-cloud-connectors/
1. **Spring Cloud Connectors**: [Spring Cloud Connectors][spring-cloud-connectors] are used by the Spring Music app to detect things like bound CF services. They work seamlessly with Kf.

1. **Configuring the Java version**: You will specify the version of Java you
   want the buildpack to use.

## Prerequisites

### Install Kf
[install-instructions]: ../install/
Follow the [installation instructions][install-instructions] to create a new
Kubernetes cluster with Kf installed.

### Minibroker
[minibroker]: ../minibroker/
Follow [these instructions][minibroker] to install the Minibroker service broker
into your cluster. Minibroker will allow you to provision a Postgresql database
and configure your application to use it.

To confirm that Minibroker is installed and available to your cluster, run `kf marketplace` and you should see output similar to:

```sh
$ kf marketplace
5 services can be used in namespace "demo", use the --service flag to list the plans for a service

BROKER      NAME        NAMESPACE  STATUS  DESCRIPTION
minibroker  mariadb                Active  Helm Chart for mariadb
minibroker  mongodb                Active  Helm Chart for mongodb
minibroker  mysql                  Active  Helm Chart for mysql
minibroker  postgresql             Active  Helm Chart for postgresql
minibroker  redis                  Active  Helm Chart for redis
```

## Deploy

### Clone source
[spring-music-source]: https://github.com/cloudfoundry-samples/spring-music/archive/fe15454c3b285bb8bdcef1c1b63252ad0d2da923.zip
1. Download and extract the [Spring Music source][spring-music-source]

    ```sh
    wget https://github.com/cloudfoundry-samples/spring-music/archive/fe15454c3b285bb8bdcef1c1b63252ad0d2da923.zip
    unzip fe15454c3b285bb8bdcef1c1b63252ad0d2da923.zip
    cd spring-music-*
    ```

1. Edit `manifest.yml`, removing the `path` key and adding an environment
   variable that will make the build use Java 8. Your resulting `manifest.yml`
   should look like:

    ```sh
    ---
    applications:
    - name: spring-music
      memory: 1G
      random-route: true
      env:
        BP_JAVA_VERSION: 8.*
    ```

### Push app

[create-space]: /docs/install.md#create-and-target-a-space
1. Deploy (this assumes you've already `kf target`'d a space; see [these
   docs][create-space] for more detail):

    ```sh
    kf push spring-music
    ```

1. Use the proxy feature to access the deployed app, then load `http://localhost:8080` in your browser:

    ```sh
    kf proxy spring-music
    ```

[ss1]: sm1.png
    ![Screenshot 1][ss1]

    > The deployed app includes a UI element showing which (if any) Spring profile is being used. No profile is being used here, indicating an in-memory database is in use.

### Create and bind database

1. Create a Postgres service via the Minibroker installed in the marketplace:

    ```sh
    kf create-service postgresql 11-4-0 spring-music-db -c '{"postgresqlDatabase":"smdb", "postgresDatabase":"smdb"}'
    ```

1. Bind the service instance to the Spring Music app:

    ```sh
    kf bind-service spring-music spring-music-db -c '{"postgresqlDatabase":"smdb", "postgresDatabase":"smdb"}'
    ```

1. Inject the bindings into the app's env:

    ```sh
    kf set-env spring-music VCAP_SERVICES "`kf vcap-services spring-music`"
    ```

1. Run `kf env spring-music` and verify that `VCAP_SERVICES` is set. It should
   look similar to:

    ```sh
    NAME             VALUE
    BP_JAVA_VERSION  8.*
    VCAP_APPLICATION
    {"application_name":"spring-music","name":"spring-music","space_name":"demo"}
    VCAP_SERVICES
    {"postgresql":[{"binding_name":"spring-music-db","instance_name":"spring-music-db","name":"kf-binding-spring-music-spring-music-db","label":"postgresql","tags":null,"plan":"11-4-0","credentials":{"Protocol":"postgresql","database":"smdb","host":"honorary-snail-postgresql.demo.svc.cluster.local","password":"***","port":"5432","postgresql-password":"***","uri":"postgresql://postgres:***@honorary-snail-postgresql.demo.svc.cluster.local:5432/smdb","username":"postgres"}}]}
    ```

1. (Optional) View the binding details:

    ```sh
    kf bindings
    ```

1. `kf proxy` to the app again and view it in your web browser. The Spring profile should be shown, indicating the Postgres service you created and bound is being used:

[ss2]: sm2.png
    ![Screenshot 2][ss2]

## Destroy

1. Unbind and delete the Postgres service:

    ```sh
    kf unbind-service spring-music spring-music-db
    kf delete-service spring-music-db
    ```

1. Delete the app:

    ```sh
    kf delete spring-music
    ```

