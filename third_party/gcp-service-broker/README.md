# GCP Service Broker Helm chart

This is a helm chart for deploying the GCP Service Broker into Kubernetes.

## Configuration

By default, the helm chart provides enough to get you running, but you'll need
to change things for a robust production environment.

For all environments:

* Set `broker.service_account_json` to be the JSON key of a service account that
  has owner permission on the GCP project you want to manage.
* Add key/value pairs to `broker.env` for environment variables you want to set
  on the broker. You can find a list of possible environment variables in the
  [customization document](https://github.com/GoogleCloudPlatform/gcp-service-broker/blob/master/docs/customization.md).

Things to change for production:

* `mysql.embedded` should be set to false and you should provide credentials
  for an external MySQL database that has automatic backups and failover.

## Building

You can build a copy of this Helm chart suitable for releasing with the following commands:

``` .sh
BROKER_VERSION=5.0.0
helm package --app-version=$BROKER_VERSION --dependency-update --version=$BROKER_VERSION .
```

## Installing

| Tutorial Name | Tutorial Link |
|:--------------|:--------------|
| Install the Service Broker into K8s for use with CF | [![Open in Cloud Shell](http://gstatic.com/cloudssh/images/open-btn.svg)](https://console.cloud.google.com/cloudshell/editor?cloudshell_git_repo=https%3A%2F%2Fgithub.com%2FGoogleCloudPlatform%2Fgcp-service-broker&cloudshell_open_in_editor=values.yaml&cloudshell_working_dir=deployments%2Fhelm%2Fgcp-service-broker&cloudshell_tutorial=cf-tutorial.md&cloudshell_git_branch=develop) |
