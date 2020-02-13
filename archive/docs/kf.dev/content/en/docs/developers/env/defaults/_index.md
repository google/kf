---
title: "Default Environment Variables"
linkTitle: "Default Environment Variables"
weight: 30
description: >
  Learn about the default environment variables available to an application in
  Kf and how they compare to CF.
---

## Compared to CF

The following table lists default environment variables available to an
application deployed to Cloud Foundry and includes a column indicating if that
variable is available by default in Kf.

| Name                    | Available in Kf? |
| ---                     | ---              |
| VCAP_SERVICES           | ✔️                |
| VCAP_APPLICATION        | ✔️                |
| VCAP_APP_HOST           | 🚫               |
| VCAP_APP_PORT           | 🚫               |
| CF_INSTANCE_ADDR        | 🚫               |
| CF_INSTANCE_CERT        | 🚫               |
| CF_INSTANCE_GUID        | 🚫               |
| CF_INSTANCE_INDEX       | 🚫               |
| CF_INSTANCE_INTERNAL_IP | 🚫               |
| CF_INSTANCE_IP          | 🚫               |
| CF_INSTANCE_KEY         | 🚫               |
| CF_INSTANCE_PORT        | 🚫               |
| CF_INSTANCE_PORTS       | 🚫               |
| CF_SYSTEM_CERT_PATH     | 🚫               |
