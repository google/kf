---
title: "Tasks Overview"
linkTitle: "Overview"
description: > 
    Understand how tasks work in Kf.
weight: 1
---

## About Tasks

Unlike Apps which run indefinitely and restart if the process terminates, Tasks run a process until it completes and then stop.
Tasks are run in their own containers and are based on the configuration and binary of an existing App.

Tasks are not accessible from routes, and should be used for one-off or scheduled recurring work necessary for the health of an
application.

## Use cases for Tasks

*  Migrating a database
*  Running a batch job (scheduled/unscheduled)
*  Sending an email
*  Transforming data (ETL)
*  Processing data (upload/backup/download)

## How Tasks work

Tasks are executed asynchronously and run independently from the parent App or other Tasks running on the same App. An App created for running Tasks does not have routes created or assigned, and the **Run** lifecycle is skipped. The **Source code upload** and **Build** lifecycles still proceed and result in a container image used for running Tasks after pushing the App (see App lifecycles at [Deploying an Application]({{<relref "deploying-an-app">}})).

The lifecycle of a Task is as follows:

1. You push an App for running tasks with the `kf push APP_NAME --task` command.
2. You run a Task on the App with the `kf run-task APP_NAME` command. Task inherits the environment variables, service bindings, resource allocation, start-up command, and security groups bound to the App.
3. Kf creates a Tekton [PipelineRun](https://github.com/tektoncd/pipeline/blob/master/docs/pipelineruns.md) with values from the App and parameters from the `run-task` command.
4.  The Tekton PipelineRun creates a Kubernetes Pod which launches a container based on the configurations on the App and Task.
5.  Task execution stops (Task exits or is terminated manually), the underlying Pod is either stopped or terminated. Pods of stopped Tasks are preserved and thus Task logs are accessible via the `kf logs APP_NAME --task` command.
6.  If you terminate a Task before it stops, the Tekton PipelineRun is cancelled (see [Cancelling a PipelineRun](https://github.com/tektoncd/pipeline/blob/master/docs/pipelineruns.md#cancelling-a-pipelinerun)), the underlying Pod together with the logs are deleted. The logs of termianted Tasks are delivered to the cluster level logging streams if configured (e.g. Stackdriver, Fluentd).
7.  If the number of Tasks run on an App is greater than 500, the oldest Tasks are automatically deleted.


## Tasks retention policy

Tasks are created as custom resources in the Kubernetes cluster, therefore, it is important not to exhaust the space of the underlying `etcd` database. By default, Kf only keeps the latest 500 Tasks per each App. Once the number of Tasks reach 500, the oldest Tasks (together with the underlying Pods and logs) will be automatically deleted.

## Task logging and execution history

Any data or messages the Task outputs to STDOUT or STDERR is available by using the `kf logs APP_NAME --task` command. Cluster level logging mechanism (such as Stackdriver, Fluentd) will deliver the Task logs to the configured logging destination.


## Scheduling Tasks

As described above, Tasks can be run asynchronously by using the `kf run-task APP_NAME` command.
Alternatively, you can schedule Tasks for execution by first creating a Job using
the `kf create-job` command, and then scheduling it with the
`kf schedule-job JOB_NAME` command. You can schedule that Job to automatically
run Tasks on a specified [unix-cron](https://man7.org/linux/man-pages/man5/crontab.5.html) schedule.


### How Tasks are scheduled

Create and schedule a Job to run the Task. A Job describes the Task to execute
and automatically manages Task creation.

Tasks are created on the schedule even if previous executions of the Task are still running.
If any executions are missed for any reason, only the most recently missed execution
are executed when the system recovers.

Deleting a Job deletes all associated Tasks. If any associated Tasks were still
in progress, they are forcefully deleted without running to completion.

Tasks created by a scheduled Job are still subject to the
[Task retention policy](#tasks_retention_policy).


### Differences from PCF Scheduler

PCF Scheduler allows multiple schedules for a single Job while Kf
only supports a single schedule per Job. You can replicate the PCF Scheduler
behavior by creating multiple Jobs, one for each schedule.

