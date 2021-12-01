---
title: Event types
---

Eventing supports both [Cloud Events](https://cloudevents.io/) and legacy events. Event Publisher Proxy converts legacy events to Cloud Events and adds the `sap.kyma.custom` prefix.

For a Subscription Custom Resource, the fully qualified event type takes the sample form of `sap.kyma.custom.commerce.order.created.v1` or `sap.kyma.custom.commerce.Account.Root.Created.v1`.

The event type is composed of the following components:
- Prefix: `sap.kyma.custom`
- Application: `commerce`
- Event: can have two or more segments separated by `.`; for example, `order.created` or `Account.Root.Created`
- Version: `v1`

For publishers, the event type takes this sample form:
- `order.created` or `Account.Root.Created` for legacy events coming from the `commerce` application
- `sap.kyma.custom.commerce.order.created.v1` or `sap.kyma.custom.commerce.AccountRoot.Created.v1` for Cloud Events


## Event name cleanup

In some cases, Eventing needs to modify the event name before dispatching an event. This is done in order to conform to Cloud Event specifications.

- If the event contains more than two segments, Eventing combines them into two segments when creating the underlying Eventing infrastructure. For example, `Account.Root.Created` becomes `AccountRoot.Created`.

- In case the Application name contains `-` or `.`, the underlying Eventing services uses a clean name with alphanumeric characters only. (For example, `system-prod` becomes `systemprod`).
This could lead to a naming collision. For example, both `system-prod` and `systemprod` could become `systemprod`.
A solution for this is to provide an `application-type` label (with alphanumeric characters only) which is then used by the Eventing services instead of the Application name. If the `application-type` label also contains `-` or `.`, the underlying Eventing services clean it and use the cleaned label.
