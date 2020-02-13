---
title: "Troubleshooting"
linkTitle: "Troubleshooting"
weight: 30
description: >
  Learn how to troubleshoot common error messages you may encounter when
  deploying or running an application with Kf.
---

## Executable JARs
[executable-jar]: https://docs.spring.io/spring-boot/docs/current/reference/html/deployment-install.html 
If you encounter the error message `zip: not a valid zip file` while deploying
your Spring application, you may be attempting to deploy an [executable
JAR][executable-jar]. Executable JARs are not supported by Kf's build system.
Refer to Spring's documentation on [executable JARs][executable-jar] to disable it
in your application, then attempt to deploy the standard, non-executable JAR
with Kf again.

[exe-jar-issue]: https://github.com/google/kf/issues/579
You may refer to [kf/579][exe-jar-issue] for more background on this issue.
