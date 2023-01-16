---
title: Event names
---

Event names depend on the type of event. Eventing supports the following event types:
- [CloudEvents](https://cloudevents.io/) - they use a specification for describing event data in a common way.
- legacy events - they are converted to CloudEvents by [Event Publisher Proxy](./00-architecture/evnt-01-architecture.md#event-publisher-proxy).

## Event name format

For a Subscription Custom Resource, the fully qualified event name takes the sample form of `order.created.v1` or `Account.Root.Created.v1`.

The event type is composed of the following components:
- Event: can have two or more segments separated by `.`; for example, `order.created` or `Account.Root.Created`
- Version: `v1`

For publishers, the event type takes this sample form:
- `order.created` or `Account.Root.Created` for legacy events coming from the `commerce` application
- `order.created.v1` or `Account.Root.Created.v1` for CloudEvents

## Event name cleanup

To conform to Cloud Event specifications, sometimes Eventing must modify the event name before dispatching an event.

### Special characters

If the event name contains any prohibited characters as per [NATS JetStream specifications](https://docs.nats.io/running-a-nats-service/nats_admin/jetstream_admin/naming), the underlying Eventing services use a clean name with allowed characters only; for example, `system>prod*` becomes `systemprod`.

This can lead to a naming collision. For example, both `system>prod` and `systemprod` become `systemprod`. While this doesn't result in an error, it can cause Eventing to not work as expected. Take a look into this [troubleshooting guide](../04-operation-guides/troubleshooting/eventing/evnt-03-type-collision.md) for more information.
