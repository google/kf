# `spring-boot-cnb`
The Cloud Foundry Spring Boot Buildpack is a Cloud Native Buildpack V3 that runs Spring Boot applications.

## Detection
The detection phase passes if:

* The build plan contains `jvm-application`

## Build
If the build plan contains

* `jvm-application`
  * Checks for the existence of a `Spring-Boot-Version` manifest key
  * If found,
    * Contributes suitably configured process types to layers marked build, cache, and launch
  * Checks for the existence of `.groovy` files, all of which must be `POGO` or configuration files
  * If found,
    * Contributes the `spring-boot-cli` binary and suitably configured process types to a layer marked launch

## License
This buildpack is released under version 2.0 of the [Apache License][a].

[a]: https://www.apache.org/licenses/LICENSE-2.0
