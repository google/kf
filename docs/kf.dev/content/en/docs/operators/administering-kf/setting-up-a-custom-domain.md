---
title: "Setting up a custom domain"
weight: 55
type: "docs"
---

By default, Kf apps are accessible on domains exposed by Knative Serving.
Kf extends Knative Serving's domain system to allow path based routing, per-app domains, wildcard hosts, and other routing services.

## Set up a domain

Follow the [Knative Serving](https://knative.dev/docs/serving/using-a-custom-domain/) docs to set up a custom domain for your cluster.

## Assign domains to spaces

You can assign domains and sub-domains to each space for developers to use.
If you created spaces before setting up a domain, then those spaces will have Knative Serving's placeholder domain by default.
If you create spaces after setting up a domain, they will have the Knative Serving default domain you specified.

Use `kf space` to view the domains assigned to a space:

```
$ kf space my-space
# snip
Execution:
  Environment:
    ENVIRONMENT:  production
  Domains:
    Name                     Default?
    my-space.example.com     true
```

You can also use `kf configure-space get-domains` for a YAML view:

```
$ kf configure-space get-domains my-space
- default: true
  domain: my-space.example.com
```

Assign a new domain using `kf configure-space append-domain`:

```
$ kf configure-space append-domain my-space myspace.mycompany.com
# diff printed
```

Then make it the default with `kf configure-space set-default-domain`:

```
$ kf configure-space set-default-domain my-space myspace.mycompany.com
# diff printed
```

And finally, delete the original placeholder domain:

```
$ kf configure-space remove-domain my-space my-space.example.com
# diff printed
```

## Known issues

* By default, Knative Serving uses `example.com` as a domain if none is configured. [#566](https://github.com/google/kf/issues/566)
* Knative Serving will ALWAYS make apps available on `<app-name>.<namespace>.<domain>` [#410](https://github.com/google/kf/issues/410)
