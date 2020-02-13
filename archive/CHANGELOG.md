# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.2.0] - 2019-10-18

### Added

* Support for `.cfignore` (or `.kfignore`) files
* Support for Dockerfile builds via kaniko
* Manifest fields for HTTP2 and no-start to match CLI flags so an app can be (almost) entirely declarative
* Display of cluster internal address to app and apps command
* Way for user to select kf release version via kf install
* Ability to define stacks in manifests and the CLI
* Support for the legacy `buildpack` manifest field
* `kf proxy-route` command
* Support for command property in manifest
* Forwarding headers `X-Forwarded-Host` and `Forwarded` to routes

### Changed
* `kf create-service` is now synchronous by default
* `kf delete-service` is now synchronous by default
* The `restart` command to now be synchronous by default
* The `scale` command to now be synchronous by default
* The `stop` command to now be synchronous by default
* The `start` command to now be synchronous by default
* The `set-env` command to now be synchronous by default
* The `unset-env` command to now be synchronous by default
* The `map-route` command to now be synchronous by default
* The `unmap-route` command to now be synchronous by default
* Default app deletion behavior to be synchronous
* The `--grpc` flag to `--enable-http2` to reflect what's actually happening in Knative Serving
* Changed default YAML parser to be `sigs.k8s.io/yaml`

### Fixed
* Installer using deprecated GKE version
* The warning about unofficial fields to show all Kf added fields
* Release script to properly login and setup GOPATH

## [0.1.0] - 2019-09-06

### Added

* Various sub-commands to replicate the CF push flow
