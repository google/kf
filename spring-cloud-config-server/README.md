# Spring Cloud Config Server example

This directory contains an example implementation of the Spring Cloud Config
Server that can be pushed to Kf.

The configuration server reads configuration from a backing data store e.g. git,
Hashicorp Vault, CredHub, or the local file system and exposes it over HTTP.
This allows client applications to decouple configuration from source code which
reduces risk during application deployments and configuration rollout.

## Configuring

The application uses the `GIT_URI` listed in the `manifest.yaml` file as the
source of the configuration.

See the [reference documentation][config-reference] for full configuration
options that can be set in `src/main/resources/application.properties`.

## Using in production

If you're going to use this application in production, you should:

* Fork the code.
* Update the Java version to match the one in your environment.
* Use your organization's parent POM (if applicable).
* Add authentication suitable for your organization.
* Increase the number of instances running for high availability.
* Decide if you want the App to only have internal routes.
* Change the amount of RAM, CPU, and disk attached.
* Open additional Spring Boot Actuator endpoints used in your environment.

## Endpoints

The configuration server exposes the following endpoints:

* `/actuator/health` - Exposes a health indicator.
* `/{application}/{profile}` - Exposes the configuration of an application for
  a specific environment as JSON.
* `/{application}-{profile}.yml` - Exposes the configuration of an application
  for a specific environment as YAML.
* `/{application}-{profile}.properties` - Exposes the configuration of an
  application for a specific environment as a Java properties file.

A full list of endpoints is available in the
[reference documentation][config-reference].

[config-reference]:https://cloud.spring.io/spring-cloud-config/reference/html/
