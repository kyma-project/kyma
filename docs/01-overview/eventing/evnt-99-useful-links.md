---
title: Useful links
---

Find out more about Eventing by following the links on this page.

To learn more about how Eventing works, see:

- [Eventing architecture](../../05-technical-reference/00-architecture/evnt-01-architecture.md) - describes how Eventing works and the main actors involved, such as the Eventing Controller and Event Publisher Proxy.
- [Event names](../../05-technical-reference/evnt-01-event-names.md) - contains information about event names and event name cleanup.
- [EventingBackend CR](../../05-technical-reference/00-custom-resources/evnt-02-eventingbackend.md) - describes the EventingBackend custom resource, which shows the current status of Kyma Eventing.
- [Subscription CR](../../05-technical-reference/00-custom-resources/evnt-01-subscription.md) - describes the Subscription custom resource, which you need to subscribe to events.
- [CloudEvents](https://cloudevents.io/) - provides information about the CloudEvents specification used in Kyma.
- [NATS JetStream](https://docs.nats.io/nats-concepts/jetstream) - provides more information about the backend technology behind Eventing in Kyma. [Eventing Architecture](../../05-technical-reference/00-architecture/evnt-01-architecture.md#jet-stream) provides details on the new functionalities and higher qualities of service on top of Core NATS.

To perform tasks with Eventing, go through these tutorials:

- [Tutorial: Trigger your workload with an event](../../02-get-started/04-trigger-workload-with-event.md) - part of the [Get Started guides](../../02-get-started), shows how to deploy a Function and trigger it with an event.
- [Tutorial: Create Subscription subscribing to multiple event types](../../03-tutorials/00-eventing/evnt-02-subs-with-multiple-filters.md) - shows how to subscribe to one or more event types using the Kyma Subscription.
- [Tutorial: Event name cleanup in Subscriptions](../../03-tutorials/00-eventing/evnt-03-type-cleanup.md) - explains how Kyma Eventing filters out prohibited characters from event names.
- [Tutorial: Changing Events Max-In-Flight in Subscriptions](../../03-tutorials/00-eventing/evnt-04-change-max-in-flight-in-sub.md) - shows how to set idle "in-flight messages" limit in Kyma Subscriptions.
- [Tutorial: Publish legacy events using Kyma Eventing](../../03-tutorials/00-eventing/evnt-05-send-legacy-events.md) - demonstrates how to send legacy events using Kyma Eventing.

To troubleshoot Eventing-related issues:
- [Basic Eventing Troubleshooting](../../04-operation-guides/troubleshooting/eventing/evnt-01-eventing-troubleshooting.md)
- [NATS JetStream Troubleshooting](../../04-operation-guides/troubleshooting/eventing/evnt-02-jetstream-troubleshooting.md)
- [Event Type Collision](../../04-operation-guides/troubleshooting/eventing/evnt-03-type-collision.md)
- [Eventing Backend Storage Full](../../04-operation-guides/troubleshooting/eventing/evnt-04-free-jetstream-storage.md)

For other technical resources, check out these links on the Kyma GitHub repository:

- [Eventing Helm chart](https://github.com/kyma-project/kyma/tree/main/resources/eventing)
- [Event Publishing Proxy](https://github.com/kyma-project/kyma/tree/main/components/event-publisher-proxy)
- [Eventing Controller](https://github.com/kyma-project/kyma/tree/main/components/eventing-controller)
- [Grafana Dashboards for Eventing](../../04-operation-guides/operations/evnt-01-eventing-dashboards.md)
- [Eventing Metrics](../../04-operation-guides/operations/evnt-02-eventing-metrics.md)
