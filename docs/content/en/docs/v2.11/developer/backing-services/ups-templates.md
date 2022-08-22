---
title: User Provided Service Templates
description: Learn how to set up user provided services for MySQL, Redis, and RabbitMQ.
weight: 100
---

{{< note >}}
You can leverage services that aren't listed in the marketplace by creating user-provided service instances that an App can bind to. Learn more about [user-provided services]({{< relref "user-provided-services" >}}).
{{< /note >}}

## Before you begin

* Ensure your service is running and accessible on the same network running your Kf cluster.
* Ensure you have targeted the Space where you want to create the service.

## Create the user-provided instance

The following examples use the most common parameters used by applications to autoconfigure services.
Most libraries use tags to find the right bound service and a URI to connect.

### MySQL

MySQL libraries usually expect the tag `mysql` and the following parameters:

`uri`
: Example `mysql://username:password@host:port/dbname`. The [MySQL documentation](https://dev.mysql.com/doc/refman/8.0/en/connecting-using-uri-or-key-value-pairs.html) can help with creating a URI string. The port is usually `3306`.

`username`
: The connection username, required by some libraries even if included in `uri`.

`password`
: The connection password, required by some libraries even if included in `uri`.


```sh
kf cups service-instance-name \
  -p '{"username":"my-username", "password":"my-password", "uri":"mysql://my-username:my-password@mysql-host:3306/my-db"}' \
  -t "mysql"
```


### RabbitMQ

RabbitMQ libraries usually expect the tag `rabbitmq` and the following parameters:

`uri`
: Example `amqp://username:password@host:port/vhost?query`. The [RabbitMQ documentation](https://www.rabbitmq.com/uri-spec.html) can help with creating a URI string. The port is usually `5672`.

Example:

```sh
kf cups service-instance-name \
  -p '{"uri":"amqp://username:password@host:5672"}' \
  -t "rabbitmq"
```

### Redis

Redis libraries usually expect the tag `redis` and the following parameters:

`uri`
: Example `redis://:password@host:port/uery`. The [IANA Redis URI documentation](https://www.iana.org/assignments/uri-schemes/prov/redis) can help with creating a URI string. The port is usually `6379`.

Example for Redis with no AUTH configured:

```sh
kf cups service-instance-name \
  -p '{"uri":"redis://redis-host:6379"}' \
  -t "redis"
```

Example for Redis with AUTH configured:

```sh
kf cups service-instance-name \
  -p '{"uri":"redis://:password@redis-host:6379"}' \
  -t "redis"
```

## Bind your App

Once the user-provided service has been created, you can bind your App to the 
user provided service by name, just like a managed service:

```sh
kf bind-service application-name service-instance-name
```

## What's next

* Learn about [how the credentials are injected into your app]({{<relref "app-runtime#vcapservices">}}).
