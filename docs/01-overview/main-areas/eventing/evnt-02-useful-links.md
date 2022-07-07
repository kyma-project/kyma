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
- [Tutorial: Create Subscription subscribing to multiple event types](../../../03-tutorials/00-eventing/evnt-02-subs-with-multiple-filters.md) - shows how to subscribe to one or more event types using the Kyma Subscription.
- [Tutorial: Event name cleanup in Subscriptions](../../../03-tutorials/00-eventing/evnt-03-type-cleanup.md) - explains how Kyma Eventing filters out non-alphanumeric character from event names.
- [Tutorial: Changing Events Max-In-Flight in Subscriptions](../../../03-tutorials/00-eventing/evnt-04-change-max-in-flight-in-sub.md) - shows how to set idle "in-flight messages" limit in Kyma Subscriptions.
- [Tutorial: Publish legacy events using Kyma Eventing](../../../03-tutorials/00-eventing/evnt-05-send-legacy-events.md) - demonstrates how to send legacy events using Kyma Eventing.

For other technical resources, check out these links on the Kyma GitHub repository:

- [Eventing Helm chart](https://github.com/kyma-project/kyma/tree/main/resources/eventing)
- [Event Publishing Proxy](https://github.com/kyma-project/kyma/tree/main/components/event-publisher-proxy)
- [Eventing Controller](https://github.com/kyma-project/kyma/tree/main/components/eventing-controller)
