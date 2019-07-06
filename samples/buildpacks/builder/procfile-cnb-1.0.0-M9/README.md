# `procfile-cnb`
The Cloud Foundry Procfile Buildpack is a Cloud Native Buildpack V3 that enables the use of a procfile to declare process types.

## Detection
The detection phase passes if

* `Procfile` exists

## Build
If the build plan contains

* `procfile`
  * Contributes process types declared in the procfile to `launch.toml`

## License
This buildpack is released under version 2.0 of the [Apache License][a].

[a]: https://www.apache.org/licenses/LICENSE-2.0
