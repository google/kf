---
title: Use managed services
weight: 50
description: >
  Learn to use the marketplace to find a service, create it, and bind it to an app.
---

## Find a service

Use the `kf marketplace` command to find a service you want to use in your App.
Running the command without arguments will show all the service classes
available. A **service class** represents a specific type of service e.g. a
MySQL database or a Postfix SMTP relay.

```
$ kf marketplace
5 services can be used in Space "test", use the --service flag to list the plans for a service

Broker      Name        Space      Status  Description
minibroker  mariadb                Active  Helm Chart for mariadb
minibroker  mongodb                Active  Helm Chart for mongodb
minibroker  mysql                  Active  Helm Chart for mysql
minibroker  postgresql             Active  Helm Chart for postgresql
minibroker  redis                  Active  Helm Chart for redis
```

Service classes can have multiple plans available. A **service plan** generally
corresponds to a version or pricing tier of the software. You can view the plans
for a specific service by supplying the service name with the marketplace
command:

```
$ kf marketplace --service mysql
Name    Free  Status  Description
5-7-14  true  Active  Fast, reliable, scalable, and easy to use open-source relational database system.
5-7-27  true  Active  Fast, reliable, scalable, and easy to use open-source relational database system.
5-7-28  true  Active  Fast, reliable, scalable, and easy to use open-source relational database system.
```

## Provision a service

Once you have identified a service class and plan to provision, you can create
an instance of the service using `kf create-service`:

```
$ kf create-service mysql 5-7-28 my-db
Creating service instance "my-db" in Space "test"
Waiting for service instance to become ready...
Success
```

Services are provisioned into a single Space. You can see the services in the
current Space by running `kf services`:

```
$ kf services
Listing services in Space: "test"
Name   ClassName  PlanName  Age   Ready  Reason
my-db  mysql      5-7-28    111s  True   <nil>
```

You can delete a service using `kf delete-service`:

```
$ kf delete-service my-db
```

## Bind a service

Once a service has been created, you can **bind** it to an App, which will
inject credentials into the App so the service can be used. You can create
the binding using `kf bind-service`:

```
$ kf bind-service my-app my-db
Creating service instance binding "binding-my-app-my-db" in Space "test"
Waiting for service instance binding to become ready...
Success
```

You can list all bindings in a Space using the `kf bindings` command:

```
$ kf bindings
Listing bindings in Space: "test"
Name                  App     Service  Age  Ready
binding-my-app-my-db  my-app  my-db    82s  True
```

Once a service is bound, restart the App using `kf restart` and the credentials
will be in the [`VCAP_SERVICES`]({{<relref "app-runtime#vcapservices">}}) environment variable.

You can delete a service binding with the `kf unbind-service` command:

```
$ kf unbind-service my-app my-db
```
