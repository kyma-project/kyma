---
title: Eventing
type: Eventing
---

Eventing in Kyma can be used to send and receive events from applications. For example, you can subscribe to events from an external application, and when the user performs an action there, you can trigger your Function or microservice. To subscribe to events, you need to use the Kyma [Subscription CRD](../technical-reference/subtab_customresources/eventing-event-subscription.md).

Kyma supports [Cloud Events](https://cloudevents.io/) - a common specification for describing event data - and legacy events. Legacy events are converted to Cloud Events by Kyma.

## Further reading
To learn more about Eventing, see these documents:

- [Eventing architecture](../technical-reference/subtab_architecture/arch-eventing-01.md) - describes how Eventing works and the main actors involved such as the Eventing Controller and Event Publisher Proxy. You can also find more information about event types here.
- [Event processing and delivery](../technical-reference/subtab_architecture/arch-eventing-event-processing.md) - contains a diagram and explanation of the the Eventing processing and delivery flow.
- [Subscription CRD](../technical-reference/subtab_customresources/eventing-event-subscription.md) - describes the `Subscription` CRD which you need to subscribe to events.
- [Tutorial: Send events without a Kyma Application](../tutorials/tut-send-events-without-kyma-app.md) - explains how to send events without the need for a Kyma Application.

> **NOTE:** The two tutorials below are part of the Getting Started Guide and are not meant to be standalone. As a prerequisite you should complete the previous tutorials first, especially [Connect an external application](docs/get-started/08-connect-external-application.md).
- [Tutorial: Trigger microservice with event](../get-started/09-trigger-microservice-with-event.md)
- [Tutorial: Trigger function with event](../get-started/13-trigger-function-with-event.md)