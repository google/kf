---
title: Configure user-provided services
weight: 80
description: >
  Learn how to inject existing services into your app.
---

Users can leverage services that aren't available in the marketplace by creating user-provided service instances.
Once created, user-provided service instances behave like managed service instances created through `kf create-service`.
Creating, listing, updating, deleting, binding, and unbinding user-provided services are all supported in Kf.

## Create a user-provided service instance

{{<tip>}}The alias for `kf create-user-provided-service` is `kf cups`.{{</tip>}}

The name given to a user-provided service must be unique across all service instances in a Space, including services created through a service broker.

### Deliver service credentials to an app

A user-provided service instance can be used to deliver credentials to an App. For example, a database admin can have a set of credentials for an existing database managed outside of Kf, and these credentials include the URL, port, username, and password used to connect to the database.

The admin can create a user-provided service with the credentials and the developer can bind the service instance to the App. This allows the credentials to be shared without ever leaving the platform. Binding a service instance to an App has the same effect regardless of whether the service is a user-provided service or a marketplace service.

The App is configured with the credentials provided by the user, and the App runtime environment variable [`VCAP_SERVICES`]({{<relref "app-runtime#vcapservices">}}) is populated with information about all of the bound services to that App.

A user-provided service can be created with credentials and/or a list of tags.

```sh
kf cups my-db -p '{"username":"admin", "password":"test123", "some-int-val": 1, "some-bool": true}' -t "comma, separated, tags"
```

This will create the user-provided service instance `my-db` with the provided credentials and tags. The credentials passed in to the `-p` flag must be valid JSON (either inline or loaded from a file path).

To deliver the credentials to one or more Apps, the user can run `kf bind-service`.

Suppose we have an App with one bound service, the user-provided service `my-db` defined above. The `VCAP_SERVICES` environment variable for that App will have the following contents:

```json
{
  "user-provided": [
    {
      "name": "my-db",
      "instance_name": "my-db",
      "label": "user-provided",
      "tags": [
        "comma",
        "separated",
        "tags"
      ],
      "credentials": {
        "username": "admin",
        "password": "test123",
        "some-int-val": 1,
        "some-bool": true
      }
    }
  ]
}
```

## Update a user-provided service instance


{{<tip>}}The alias for `kf update-user-provided-service` is `kf uups`.{{</tip>}}

A user-provided service can be updated with the `uups` command. New credentials and/or tags passed in completely overwrite existing ones. For example, if the user created the user-provided service `my-db` above, called `kf bind-service` to bind the service to an App, then ran the command.

```sh
kf uups my-db -p '{"username":"admin", "password":"new-pw", "new-key": "new-val"}'
```

The updated credentials will only be reflected on the App after the user unbinds and rebinds the service to the App. No restart or restage of the App is required. The updated `VCAP_SERVICES` environment variable will have the following contents:

```
{
  "user-provided": [
    {
      "name": "my-db",
      "instance_name": "my-db",
      "label": "user-provided",
      "tags": [
        "comma",
        "separated",
        "tags"
      ],
      "credentials": {
        "username": "admin",
        "password": "new-pw",
        "new-key": "new-val"
      }
    }
  ]
}
```

The new credentials overwrite the old credentials, and the tags are unchanged because they were not specified in the update command.
