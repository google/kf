---
title: "Java"
linkTitle: "Java"
weight: 10
description: >
  Learn how to use the Java buildpacks to deploy your Java apps.
---

The Java buildpack extracts libraries, builds using Maven or Gradle, optionally
configures Spring, embeds Tomcat if necessary, and reads Procfile if provided.

## Runtime

 * OpenJDK
 * Tomcat
 * Maven
 * Gradle

## Compatible apps

Compatible apps:

 * NPM-based apps with package.json
 * Vendored apps (with node_modules included)
 * Shrinkwrapped apps (with npm-shrinkwrap.json) are untested
 * Yarn apps are not supported

## Configuration options

1. Archive expansion
  * The first step will expand any archives
  * No options are supported
2. OpenJDK
  * Contributes a JDK and JRE
  * Use `BP_JAVA_VERSION` to set a specific version. Whildcards are supported e.g. `8.*`.
3. Build System
  * Builds using Maven or Gradle
  * Use `BP_BUILT_ARTIFACT` to override the path to the built artifact.
4. JVM Application
  * Sets the container launch point to be a Java application
  * No options are supported
5. Spring Boot
  * Adds the Spring Boot CLI if the app is a Spring app
  * No options are supported
6. Tomcat
  * Add Tomcat if the app is a WAR file
  * No options are supported
7. Procfile
  * Uses a Procfile to specify an entrypoint
  * No options are supported
8. Google Stackdriver
  * Injects Stackdriver Debugger and Profiler if a binding is attached for Stackdriver.
  * No options are supported
