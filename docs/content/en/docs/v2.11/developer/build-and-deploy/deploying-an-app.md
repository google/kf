---
title: Deploy an application
description: "Guide to deploying applications with Kf."
weight: 100
---

When pushing an app (via `kf push`) to Kf, there are
three lifecycles that Kf uses to take your source code
and allow it to handle traffic:

1. Source code upload
1. Build
1. Run

## Source code upload

The first thing that happens when you `kf push` is the Kf CLI (`kf`) packages
up your directory (either current or `--path/-p`) into a container and
publishes it to the container registry configured for the Space. This is
called the source container. The Kf CLI then creates an `App` type in Kubernetes
that contains both the source image and configuration from the App manifest and
push flags.

### Ignore files during push

In many cases, you will not want to upload certain files during `kf push` (i.e., "ignore" them).
This is where a `.kfignore` (or `.cfignore`) file can be used.
Similar to a `.gitignore` file, this file instructs the Kf CLI which
files to not include in the source code container.

To create a `.kfignore` file, create a text file named `.kfignore` in the base
directory of your app (similar to where you would store the manifest
file). Then populate it with a newline delimited list of files and directories
you don't want published. For example:

```
bin
.idea
```

This will tell the Kf CLI to not include anything in the `bin` or `.idea`
directories.

Kf supports [gitignore](https://git-scm.com/docs/gitignore) style syntax.

## Build

The Build lifecycle is handled by a Tekton
[TaskRun](https://github.com/tektoncd/pipeline/blob/master/docs/taskruns.md).
Depending on the flags that you provide while pushing, it will choose a specific
Tekton [Task](https://github.com/tektoncd/pipeline/blob/master/docs/tasks.md).
Kf currently has the following Tekton Tasks:

* buildpackv2
* buildpackv3
* kaniko

Kf tracks each TaskRun as a Build. If a Build succeeds, the resulting
container image is then deployed via the Run lifecycle (described below).

More information can be found at [Build runtime]({{<relref "build-runtime">}}).

## Run

The Run lifecycle is responsible for taking a container image and creating a
[Kubernetes Deployment](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/).

It also creates:

* [Istio Virtual Services](https://istio.io/docs/reference/config/networking/virtual-service/)
* [Kubernetes Secrets](https://kubernetes.io/docs/concepts/configuration/secret/)

More information can be found at [Build runtime]({{<relref "build-runtime">}}).

## Push timeouts

Kf supports setting an environment variable to instruct the CLI to time out
while pushing apps. If set, the variables `KF_STARTUP_TIMEOUT` or
`CF_STARTUP_TIMEOUT` are parsed as a golang style duration (for example `15m`,
`1h`). If a value is not set, the push timeout defaults to 15 minutes.
