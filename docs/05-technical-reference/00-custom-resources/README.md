---
title: Custom resources
---

A custom resource (CR) is an extension to the Kubernetes API which allows you to cover use cases that are not directly covered by core Kubernetes. Here's the list of the CRs developed by Kyma to support its functionality:

| Area | Custom resource |
| ---- | -------------- |
| Application Connectivity | [Application](ac-01-application.md), [CompassConnection](ra-01-compassconnection.md) |
| API Gateway | [APIRule](apix-01-apirule.md) |
| Eventing | [Subscription](evnt-01-subscription.md), [EventingBackend](evnt-02-eventingbackend.md) |
| Istio | [Istio](oper-01-istio.md) |
| Serverless | [Function](svls-01-function.md) |
| Telemetry | [LogPipeline](telemetry-01-logpipeline.md), [LogParser](telemetry-02-logparser.md), [TracePipeline](telemetry-03-tracepipeline.md) |

 > **TIP:** For information about third-party custom resources that come together with Kyma, visit the documentation of the respective project.
