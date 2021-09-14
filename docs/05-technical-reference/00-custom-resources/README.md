---
title: Custom resources
---

A custom resource (CR) is an extension to the Kubernetes API which allows you to cover use cases that are not directly covered by core Kubernetes. Here's the list of the CRs provided by Kyma to support its functionality:

| Area | Custom resource |
| ---- | -------------- |
| Application Connectivity | Application, ApplicationMapping, EventActivation, TokenRequest, CompassConnection |
| API Exposure | APIRule |
| Eventing | EventSubscription |
| Service Management | ServiceBindingUsage, UsageKind, ClusterAddonsConfiguration, AddonsConfiguration |
| Serverless | Function, GitRepository |

 > **TIP:** For information about third-party custom resources used by Kyma, such as Prometheus, visit the documentation of the respecitve project.
