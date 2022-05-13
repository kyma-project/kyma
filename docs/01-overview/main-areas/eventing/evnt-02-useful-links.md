---
title: Useful links
---

Find out more about Eventing by following the links on this page.

To learn more about how Eventing works, see:

- [Eventing architecture](../../../05-technical-reference/00-architecture/evnt-01-architecture.md) - describes how Eventing works and the main actors involved, such as the Eventing Controller and Event Publisher Proxy.
- [Event names](../../../05-technical-reference/evnt-01-event-names.md) - contains information about event names and event name cleanup.
- [Subscription CR](../../../05-technical-reference/00-custom-resources/evnt-01-subscription.md) - describes the Subscription custom resource, which you need to subscribe to events.
- [Cloud Events](https://cloudevents.io/) - provides information about the Cloud Events specification used in Kyma.
- [NATS](https://nats.io/) - provides more information about the backend technology behind Eventing in Kyma.
- [JetStream](https://docs.nats.io/nats-concepts/jetstream) - provides details on the new functionalities and higher qualities of service on top of NATS. Read [Eventing Architecture](../../../05-technical-reference/00-architecture/evnt-01-architecture.md#jet-stream) for more information.

To perform tasks with Eventing, go through these tutorials:

- [Tutorial: Trigger your workload with an event](../../../02-get-started/04-trigger-workload-with-event.md) - part of the [Get Started guides](../../../02-get-started), shows how to deploy a Function and trigger it with an event.
- [Tutorial: Send events without a Kyma Application](../../../03-tutorials/00-eventing/evnt-02-setup-in-cluster-eventing.md) - explains how to send events without the need for a Kyma Application.
- [Tutorial: Send events from outside the Kyma cluster](../../../03-tutorials/00-eventing/evnt-01-send-event-outside-cluster.md) - shows how to send different types of events from an external Application.
- [Tutorial: Introduction to JetStream](../../../03-tutorials/00-eventing/evnt-03-introduction-to-jetstream.md) - explains the key concepts of JetStream and how to verify `at least once` delivery with JetStream.

For other technical resources, check out these links on the Kyma GitHub repository:

- [Eventing Helm chart](https://github.com/kyma-project/kyma/tree/main/resources/eventing)
- [Event Publishing Proxy](https://github.com/kyma-project/kyma/tree/main/components/event-publisher-proxy)
- [Eventing Controller](https://github.com/kyma-project/kyma/tree/main/components/eventing-controller)
