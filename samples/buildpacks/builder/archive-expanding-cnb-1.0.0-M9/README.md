# `archive-expanding-cnb`
The Cloud Foundry Archive Expanding Buildpack is a Cloud Native Buildpack V3 that expands archived (JAR, ZIP, TAR) applications before other buildpacks are presented with them.

## Detection
The detection phase passes if

* A single `.jar`, `.war`, `.tar`, `.tar.gz`, `.tgz`, or `.zip` file is detected in the root of the workspace
  * Contributes `archive-expanding` to the build plan
  * Contributes `jvm-application` to the build plan

## Build
If the build plan contains

* `archive-expanding`
  * Expands the archive into the root of the workspace
  * Remove the archive from the workspace

## License
This buildpack is released under version 2.0 of the [Apache License][a].

[a]: http://www.apache.org/licenses/LICENSE-2.0
