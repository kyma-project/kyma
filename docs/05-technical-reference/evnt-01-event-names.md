---
title: Event names
---

Event names depend on the type of event. Eventing supports the following event types:
- [Cloud Events](https://cloudevents.io/) - they use a specification for describing event data in a common way.
- legacy events - they are converted to Cloud Events by [Event Publisher Proxy](./00-architecture/evnt-01-architecture.md#event-publisher-proxy), which also adds a `sap.kyma.custom` prefix.

## Event name format

For a Subscription Custom Resource, the fully qualified event name takes the sample form of `sap.kyma.custom.commerce.order.created.v1` or `sap.kyma.custom.commerce.Account.Root.Created.v1`.

The event type is composed of the following components:
- Prefix: `sap.kyma.custom`
- Application: `commerce`
- Event: can have two or more segments separated by `.`; for example, `order.created` or `Account.Root.Created`
- Version: `v1`

For publishers, the event type takes this sample form:
- `order.created` or `Account.Root.Created` for legacy events coming from the `commerce` application
- `sap.kyma.custom.commerce.order.created.v1` or `sap.kyma.custom.commerce.Account.Root.Created.v1` for Cloud Events

## Event name cleanup

To conform to Cloud Event specifications, sometimes Eventing must modify the event name before dispatching an event.

### Events with more than two segments

If the event name contains more than two segments, Eventing combines them into two segments when creating the underlying Eventing infrastructure. For example, `Account.Root.Created` becomes `AccountRoot.Created`.

### Non-alphanumeric characters

If the Application name contains any non-alphanumeric character `[^a-zA-Z0-9]+`, the underlying Eventing services use a clean name with alphanumeric characters only `[a-zA-Z0-9]+`; for example, `system-prod` becomes `systemprod`.

This could lead to a naming collision. For example, both `system-prod` and `systemprod` become `systemprod`. While this won't result in an error, it can cause Kyma to not work as expected. Take a look into this [troubleshooting guide](../04-operation-guides/troubleshooting/eventing/evnt-03-type-collision.md) for more information.
