# `build-system-cnb`
The Cloud Foundry Build System Buildpack is a Cloud Native Buildpack V3 that enables the building of JVM applications from source.

This buildpack is designed to work in collaboration with other buildpacks.

## Detection
The detection phase passes if

* `<APPLICATION_ROOT>/build.gradle` or `<APPLICATION_ROOT>/build.gradle.kts` exists
  * Contributes `gradle` to the build plan
  * Contributes `jvm-application` to the build plan
  * Contributes `openjdk-jdk` to the build plan

* `<APPLICATION_ROOT>/pom.xml` exists
  * Contributes `maven` to the build plan
  * Contributes `jvm-application` to the build plan
  * Contributes `openjdk-jdk` to the build plan

## Build
If the build plan contains

* `gradle`
  * Contributes a layer marked `cache` and links it to `$HOME/.gradlew`
  * If `<APPLICATION_ROOT>/gradlew` exists
    * Contributes a layer marked `build`, `cache`, and `launch` by running `<APPLICATION_ROOT>/gradlew -x test build`
  * If `<APPLICATION_ROOT>/gradlew` does not exist
    * Contributes Gradle distribution to a layer marked `cache` with all commands on `$PATH`
    * Contributes a layer marked `build`, `cache`, and `launch` by running `<GRADLE_ROOT>/bin/gradle -x test build`
  * Replaces`<APPLICATION_ROOT>` with a link to compiled application layer
  * If `$BP_BUILT_ARTIFACT` exists, uses the specified path (including glob patterns) as the built artifact

* `maven`
  * Contributes a layer marked `cache` and links it to `$HOME/.m2`
  * If `<APPLICATION_ROOT>/mvnw` exists
    * Contributes a layer marked `build`, `cache`, and `launch` by running `<APPLICATION_ROOT>/mvnw -Dmaven.test.skip=true package`
  * If `<APPLICATION_ROOT>/mvnw` does not exist
    * Contributes Maven distribution to a layer marked `cache` with all commands on `$PATH`
    * Contributes a layer marked `build`, `cache`, and `launch` by running `<MAVEN_ROOT>/bin/mvn -Dmaven.test.skip=true package`
  * Replaces`<APPLICATION_ROOT>` with a link to compiled application layer
  * If `$BP_BUILT_ARTIFACT` exists, uses the specified path (including glob patterns) as the built artifact

## License
This buildpack is released under version 2.0 of the [Apache License][a].

[a]: https://www.apache.org/licenses/LICENSE-2.0

