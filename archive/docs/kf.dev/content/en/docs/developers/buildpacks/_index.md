---
title: "Buildpacks"
linkTitle: "Buildpacks"
weight: 30
description: >
  Learn how to use the default buildpacks to create Java, Node, and Go
  applications.
---

[Buildpacks](https://buildpacks.io/) are pre-defined build steps that cluster operators can expose to developers via a "Builder".
Each buildpack knows how to perform one or more operations on uploaded source code.
Buildpacks are made up of two binaries each, `detect` and `build`.

* `detect` returns true if the source code can be acted upon by the buildpack
* `build` executes some action on the code

For example, a Maven buildpack could check for the existence of a `pom.xml` file

If no buildpack is provided, Kf will run `detect` on all configured buildpacks to see which can be ran against your code.
If you provide one or more buildpacks, Kf will run them in order.

## Chains of buildpacks

A single buildpack isn't very powerful, but together they can form a complex build platform.
Multiple buildpacks can be executed together to produce complex images via buildpack **groups**.

For example, a Java buildpack group might chain the following buildpacks in order:

 * [archive-expanding](https://github.com/cloudfoundry/archive-expanding-cnb) - Expand JAR archives to classfiles if they exist
 * [openjdk](https://github.com/cloudfoundry/openjdk-cnb) Provide OpenJDK to the build process
 * [build-system](https://github.com/cloudfoundry/build-system-cnb) Run Maven or Gradle builds
 * [spring-boot](https://github.com/cloudfoundry/spring-boot-cnb) Autoconfigure Spring Boot entrypoints
 * [tomcat](https://github.com/cloudfoundry/tomcat-cnb) Provide Tomcat if the app is a WAR
 * [procfile](https://github.com/cloudfoundry/procfile-cnb/) Set an entrypoint based on the `Procfile` if one exists

## Default buildpacks

By default, Kf uses a [custom builder image](https://github.com/GoogleCloudPlatform/kf-buildpacks)
that supports building Go, Node, and Java apps.

## Known limitations

 * Kf doesn't currently support running buildpacks by URI [#619](https://github.com/google/kf/issues/619).
 * Kf does not inject the `VCAP_SERVICES` environment variable into builds [660](https://github.com/google/kf/issues/660).
