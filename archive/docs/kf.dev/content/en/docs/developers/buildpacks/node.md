---
title: "Node"
linkTitle: "Node"
weight: 30
description: >
  Learn how to use the Node buildpack to deploy your Node apps.
---

The Node buildpack gets Node, then runs `npm install` and sets the launch command to `npm start`.

## Runtime

 * Node

## Compatible apps

Compatible apps:

 * NPM-based apps with package.json
 * Vendored apps (with node_modules included)
 * Shrinkwrapped apps (with npm-shrinkwrap.json) are untested
 * Yarn apps are not supported

## Configuration options

No configuration options are supported.
