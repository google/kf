---
title: "Run Tasks"
description: > 
    Learn how to use tasks to run one-off jobs.
weight: 100
---

You can execute short-lived workflows by running them as Tasks in
Kf. Tasks are run under Apps, meaning that each Task must
have an associated App. Each Task execution uses the build artifacts from the
parent App. Because Tasks are short-lived, the App is not deployed as a
long-running application, and no routes should be created for the App or the Task.

## Push an App for running Tasks

1. Clone the [test-app repo](https://github.com/cloudfoundry-samples/test-app) repo:

    ```sh
    git clone https://github.com/cloudfoundry-samples/test-app test-app

    cd test-app
    ```

1. Push the App.

   Push the App with the `kf push APP_NAME --task` command. The `--task` flag
   indicates that the App is meant to be used for running Tasks, and thus no
   routes are created on the App, and it is not deployed as a long-running
   application:

    ```sh
    kf push test-app --task
    ```

1. Confirm that no App instances or routes were created by listing the App:

    ```sh
    kf apps
    ```

   Notice that the App is not started and has no URLs:

    ```none
    Listing Apps in Space: test-space
    Name                     Instances  Memory  Disk  CPU   URLs
    test-app                 stopped    1Gi     1Gi   100m  <nil>
    ```

## Run a Task on the App

When you run a Task on the App, you can optionally specify a start command by
using the `--command` flag. If no start command is specified, it uses the start
command specified on the App. If the App doesn't have a start command specified,
it looks up the CMD configuration of the container image. A start command must
exist in order to run the Task successfully.

```sh
kf run-task test-app --command "printenv"
```

You see something similar to this, confirming that the Task was submitted:

```none
Task test-app-gd8dv is submitted successfully for execution.
```

The Task name is automatically generated, prefixed with the App name, and
suffixed with an arbitrary string. The Task name is a unique identifier for
Tasks within the same cluster.

## Specify Task resource limits

Resource limits (such as CPU cores/Memory limit/Disk quota) can be specified in
the App (during `kf push`) or during the `kf run-task` command. The limits
specified in the `kf run-task` command take prededence over the limits specified
on the App.

To specify resource limits in an App, you can use the `--cpu-cores`,
`--memory-limit`, and `--disk-quota` flags in the `kf push` command:

```sh
kf push test-app --command "printenv" --cpu-cores=0.5 --memory-limit=2G --disk-quota=5G --task
```

To override these limits in the App, you can use the `--cpu-cores`,
`--memory-limit`, and `--disk-quota` flags in the `kf run-task` command:

```sh
kf run-task test-app --command "printenv" --cpu-cores=0.5 --memory-limit=2G --disk-quota=5G
```

## Specify a custom display name for a Task

You can optionally use the `--name` flag to specify a custom display name for a
Task for easier identification/grouping:

```none
$ kf run-task test-app --command "printenv" --name foo
Task test-app-6swct is submitted successfully for execution.

$ kf tasks test-app
Listing Tasks in Space: test space
Name              ID  DisplayName        Age    Duration  Succeeded  Reason
test-app-6swct    3   foo                1m     21s       True       <nil>
```

## Manage Tasks

View all Tasks of an App with the `kf tasks APP_NAME` command:

```none
$ kf tasks test-app
Listing Tasks in Space: test space
Name              ID  DisplayName        Age    Duration  Succeeded  Reason
test-app-gd8dv    1   test-app-gd8dv     1m     21s       True       <nil>
```

## Cancel a Task

Cancel an active Task by using the `kf terminate-task` command:

* Cancel a Task by Task name:

  ```none
  $ kf terminate-task test-app-6w6mz
  Task "test-app-6w6mz" is successfully submitted for termination
  ```

* Or cancel a Task by <code><var>APP_NAME</var></code> + Task ID:

  ```none
  $ kf terminate-task test-app 2
  Task "test-app-6w6mz" is successfully submitted for termination
  ```

{{< note >}} You can only cancel Tasks that are pending/running. Completed Tasks are not cancellable.{{< /note >}}

Cancelled Tasks have `PipelineRunCancelled` status.

```none
$ kf tasks test-app
Listing Tasks in Space: test space
Name              ID  DisplayName        Age    Duration  Succeeded  Reason
test-app-gd8dv    1   test-app-gd8dv     1m     21s       True       <nil>
test-app-6w6mz    2   test-app-6w6mz     38s    11s       False      PipelineRunCancelled
```

## View Task logs

View logs of a Task by using the `kf logs APP_NAME --task` command:

```sh
$ kf logs test-app --task
```

