# Deploy Spring Music as a Kf App

These instructions will walk you through deploying the [Cloud Foundry Spring Music][spring-music]
reference app with Kf, demonstrating a few things along the way:

1. **Building Java apps from source**: The Spring Music source will be built on
   the cluster, not locally.

1. **Service broker integration**: You will create and bind a Postgresql
   database to the Spring Music app.

1. **Spring Cloud Connectors**: [Spring Cloud Connectors][spring-cloud-connectors] are used by the Spring Music app to detect things like bound CF services. They work seamlessly with Kf.

1. **Configuring the Java version**: You will specify the version of Java you
   want the buildpack to use.

## Prerequisites
In addition to having a cluster with Kf installed, these instructions assume you have installed Minibroker as [described here][install-minibroker]. To confirm that Minibroker is installed and available to your cluster, run `kf marketplace` and you should see output similar to:

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

1. Deploy (this assumes you've already `kf target`'d a space): 

    ```sh
    kf push spring-music
    ```

1. Use the proxy feature to access the deployed app, then load `http://localhost:8080` in your browser:

    ```sh
    kf proxy spring-music
    ```

    ![Screenshot 1][ss1]

    > The deployed app includes a UI element showing which (if any) Spring profile is being used. No profile is being used here, indicating an in-memory database is in use.

1. Create a Postgres service via the Minibroker installed in the marketplace:

    ```sh
    kf create-service postgresql 11-4-0 spring-music-db -c '{"postgresqlDatabase":"smdb", "postgresDatabase":"smdb"}'
    ```

1. Bind the service instance to the Spring Music app:

    ```sh
    kf bind-service spring-music spring-music-db -c '{"postgresqlDatabase":"smdb", "postgresDatabase":"smdb"}'
    ```

1. (Optional) Verify that the binding succeeded:

    ```sh
    kf bindings
    ````

1. (Optional) Verify that the `VCAP_SERVICES` env var has been injected in your Spring Music app:

    ```sh
    kf vcap-services spring-music
    ```

1. Re-push the application so it uses the new Postgres database:

    ```sh
    kf push spring-music
    ```

1. `kf proxy` to the app again and view it in your web browser. The Spring profile should be shown, indicating the Postgres service you created and bound is being used:

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

[spring-music]: https://github.com/cloudfoundry-samples/spring-music
[spring-music-source]:
https://github.com/cloudfoundry-samples/spring-music/archive/fe15454c3b285bb8bdcef1c1b63252ad0d2da923.zip
[spring-cloud-connectors]: https://cloud.spring.io/spring-cloud-connectors/
[ss1]: sm1.png
[ss2]: sm2.png
[install-minibroker]: /docs/install.md#install-minibroker
