---
title: Useful links
---

You can find out more about Eventing by following the links on this page.

To learn more about how Eventing works, see:

- [Eventing architecture](../05-technical-reference/03-architecture/evnt-01-architecture.md) - describes how Eventing works and the main actors involved such as the Eventing Controller and Event Publisher Proxy. You can also find more information about event types here.
- [Event processing and delivery](../05-technical-reference/03-architecture/evnt-02-event-processing.md) - contains a diagram and explanation of the the Eventing processing and delivery flow.
- [Subscription CRD](../05-technical-reference/06-custom-resources/evnt-01-subscription.md) - describes the `Subscription` CRD which you need to subscribe to events.
- [Cloud Events](https://cloudevents.io/) - information about the Cloud Events specification used in Kyma
- [NATS](https://nats.io/) - learn more about the backend technology behind Eventing in Kyma

To perform tasks with Eventing:

- [Tutorial: Send events without a Kyma Application](../03-tutorials/eventing/evnt-01-setup-in-cluster-eventing.md) - explains how to send events without the need for a Kyma Application.

> **NOTE:** The two tutorials below are part of the Getting Started Guide and are not meant to be standalone. As a prerequisite you should complete the previous tutorials first, especially [Connect an external application](docs/get-started/08-connect-external-application.md).
- [Tutorial: Trigger microservice with event](../02-get-started/09-trigger-microservice-with-event.md)
- [Tutorial: Trigger function with event](../02-get-started/13-trigger-function-with-event.md)

For other technical resources, check out these links on the the Kyma GitHub repository:

- [Eventing Helm chart](https://github.com/kyma-project/kyma/tree/main/resources/eventing)
- [Event Publishing Proxy](https://github.com/kyma-project/kyma/tree/main/components/event-publisher-proxy)
- [Eventing Controller](https://github.com/kyma-project/kyma/tree/main/components/eventing-controller)

