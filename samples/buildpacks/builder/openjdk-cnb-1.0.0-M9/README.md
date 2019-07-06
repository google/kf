# `openjdk-cnb`
The Cloud Foundry OpenJDK Buildpack is a Cloud Native Buildpack V3 that provides OpenJDK JREs and JDKs to applications.

This buildpack is designed to work in collaboration with other buildpacks which request contributions of JREs and JDKs.

## Detection
The detection phase always passes and contributes nothing to the build plan, depending on other buildpacks to request
contributions.

## Build
If the build plan contains

* `openjdk-jdk`
  * Contributes a JDK to a layer marked `build` and `cache` with all commands on `$PATH`
  * If `$BP_JAVA_VERSION` is set, configures a specific version.  This value must _exactly_ match a version available in
    the buildpack so typically it would configured to a wildcard such as `8.*`.
  * Contributes `$JAVA_HOME` configured to the build layer
  * Contributes `$JDK_HOME` configure to the build layer

* `openjdk-jre`
  * Contributes a JRE to a layer with all commands on `$PATH`
  * If `$BP_JAVA_VERSION` is set, configures a specific version.  This value must _exactly_ match a version available in
    the buildpack so typically it would configured to a wildcard such as `8.*`.
  * Contributes `$JAVA_HOME` configured to the layer
  * If `metadata.build = true`
    * Marks layer as `build` and `cache`
  * If `metadata.launch = true`
    * Marks layer as `launch`

## License
This buildpack is released under version 2.0 of the [Apache License][a].

[a]: https://www.apache.org/licenses/LICENSE-2.0
