# `jvm-application-cnb`
The Cloud Foundry JVM Application Buildpack is a Cloud Native Buildpack V3 that enables the running of JVM applications.

This buildpack is designed to work in collaboration with other buildpacks.

## Detection
The detection phase passes if

* `jvm-application` exists in the build plan
  * Contributes `openjdk-jre` to the build plan
  * Contributes `openjdk-jre.metadata.launch = true` to the build plan

## Build
If the build plan contains

* `jvm-application`
  * `<APPLICATION_ROOT>/META-INF/MANIFEST.MF` with `Main-Class` key declared
    * Contributes `executable-jar` process type
    * Contributes `task` process type
    * Contributes `web` process type

## License
This buildpack is released under version 2.0 of the [Apache License][a].

[a]: https://www.apache.org/licenses/LICENSE-2.0

