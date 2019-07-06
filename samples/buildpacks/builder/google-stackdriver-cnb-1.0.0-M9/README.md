# `google-stackdriver-cnb`
The Cloud Foundry Google Stackdriver Buildpack is a Cloud Native Buildpack V3 that provides the [Google Stackdriver][g] Debugger and Profiler agents and configuration to applications.

This buildpack is designed to work in collaboration with bound service instances.

[g]: https://cloud.google.com/stackdriver/

## Detection
The detection phase passes if

* A service is bound with a payload containing `binding_name`, `instance_name`, `label`, or `tag` containing `google-stackdriver-debugger` as a substring and build plan contains `jvm-application`.
  * Contributes `google-stackdriver-debugger-java` to the build plan
* A service is bound with a payload containing `binding_name`, `instance_name`, `label`, or `tag` containing `google-stackdriver-profiler` as a substring and build plan contains `jvm-application`.
  * Contributes `google-stackdriver-profiler-java` to the build plan

## Build
If the build plan contains

* `google-stackdriver-debugger`
  * Contributes Google Stackdriver Debugger agent to a layer marked `launch`
  * Sets `-agentpath` and `com.google.cdbg.auth.serviceaccount.enable` to `$JAVA_OPTS`
  * If `$BPL_GOOGLE_STACKDRIVER_MODULE` is specified, configures the `com.google.cdbg.module` to `$JAVA_OPTS`.  Defaults to `default-module`.
  * If `$BPL_GOOGLE_STACKDRIVER_VERSION` is specified, configures the version.
  * Contributes Google Stackdriver Credentials helper to a layer marked `launch`.
    * Sets `$GOOGLE_APPLICATION_CREDENTIALS`.
* `google-stackdriver-profiler`
  * Contributes Google Stackdriver Profiler agent to a layer marked `launch`
  * Sets `-agentpath` to `$JAVA_OPTS`
  * If `$BPL_GOOGLE_STACKDRIVER_MODULE` is specified, configures the module.  Defaults to `default-module`.
  * If `$BPL_GOOGLE_STACKDRIVER_VERSION` is specified, configures the version.
  * Contributes Google Stackdriver Credentials helper to a layer marked `launch`.
    * Sets `$GOOGLE_APPLICATION_CREDENTIALS`.

## License
This buildpack is released under version 2.0 of the [Apache License][a].

[a]: https://www.apache.org/licenses/LICENSE-2.0

