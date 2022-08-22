---
title: "Schedule Tasks"
description: > 
    Learn how to schedule tasks to run periodic jobs.
weight: 200
---

You can execute short-lived workflows by running them as [Tasks]({{< relref "task" >}}).
[Running Tasks]({{< relref "run-task#push_an_app_for_running_tasks" >}})
describes how to run Tasks under Apps.

You can also schedule Tasks to run at recurring intervals
specified using the [unix-cron](https://man7.org/linux/man-pages/man5/crontab.5.html) format.
With scheduled Tasks, you first push an App running the Task as you do with an unscheduled Task,
and then create a Job to schedule the Task.

You can define a schedule so that your Task runs multiple times a day or on specific
days and months.

## Push an App for running scheduled Tasks

1. Clone the [test-app repo](https://github.com/cloudfoundry-samples/test-app):

  ```sh
  git clone https://github.com/cloudfoundry-samples/test-app test-app

  cd test-app
  ```

1. Push the App.

   Push the App with the `kf push APP_NAME --task` command. The `--task` flag indicates that the App is meant to be used for running Tasks, and thus no routes will be created on the App and it will not be deployed as a long-running application.

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

## Create a Job

To run a Task on a schedule, you must first create a Job that describes the Task:

```sh
kf create-job test-app test-job "printenv"
```

The Job starts suspended or unscheduled, and does not create Tasks until it is
manually executed by `kf run-job` or scheduled by `kf schedule-task`.

### Run a Job manually

Jobs can be run ad hoc similar to running Tasks by `kf run-task`. This option
can be useful for testing the Job before scheduling or running as needed in addition
to the schedule.

```sh
kf run-job test-job
```

This command runs the Task defined by the Job a single time immediately.

## Schedule a Job

To schedule the Job for execution, you must provide a unix-cron schedule in the
`kf schedule-job` command:

```sh
kf schedule-job test-job "* * * * *"
```

This command triggers the Job to automatically create Tasks on the specified schedule.
In this example a Task runs every minute.

You can update a Job's schedule by running `kf schedule-task` with a new schedule.
Jobs in Kf can only have a single cron schedule. This differs
from the PCF Scheduler, which allows multiple schedules for a single Job.
If you require multiple cron schedules, then you can achieve that with multiple Jobs.

## Manage Jobs and schedules

View all Jobs, both scheduled and unscheduled, in the current Space by using
the `kf jobs` command:

```none
$ kf jobs
Listing Jobs in Space: test space
Name               Schedule    Suspend  LastSchedule  Age  Ready  Reason
test-job           * * * * *   <nil>    16s           2m   True   <nil>
unscheduled-job    0 0 30 2 *  true     16s           2m   True   <nil>
```

{{< note >}} The `unscheduled-job` has a default schedule set (`0 0 30 2 *`).
This schedule is not active because the Job is suspended and is only present as a
placeholder schedule for unscheduled Jobs.{{< /note >}}

Additionally, you can view only Jobs that are actively scheduled with
the `kf job-schedules` command.

```none
$ kf job-schedules
Listing job schedules in Space: test space
Name           Schedule   Suspend  LastSchedule  Age  Ready  Reason
test-job       * * * * *  <nil>    16s           2m   True   <nil>
```

Notice how the `unscheduled-job` is not listed in the `kf job-schedules` output.

### Cancel a Job's schedule

You can stop a scheduled Job with the `kf delete-job-schedule` command:

```sh
kf delete-job-schedule test-job
```

This command suspends the Job and stops it from creating Tasks on the previous schedule.
The Job is not deleted and can be scheduled again by `kf schedule-job` to continue execution.

### Delete a Job

The entire Job can be deleted with the `kf delete-job` command:

```sh
kf delete-job test-job
```

This command deletes the Job and all Tasks that were created by the Job,
both scheduled and manual executions. If any Tasks are still running, this command
forcefully deletes them.

If you want to ensure that running Tasks are not interrupted, then first delete
the Jobs schedule with `kf delete-job-schedule`, wait for all Tasks to complete,
and then delete the job by calling `kf delete-job`.
