---
title: Eventing Architecture
---

Eventing uses Event Publisher Proxy and Eventing Controller to connect to the default NATS backend. They work together to process and deliver events in Kyma. See the [Event processing and delivery](../evnt-01-event-processing.md) document for more information.

## Event Publisher Proxy

Event Publisher Proxy receives legacy and Cloud Event publishing requests from the cluster workloads (microservices or Functions). It converts any legacy events to Cloud Events. Then, it redirects events to the NATS server. It also fetches a list of Subscriptions for a connected application.

## Eventing Controller

Eventing Controller manages the internal infrastructure in order to receive an event. It watches Subscription custom resource definitions. When an event is received in an Application, it lays down the Eventing infrastructure in NATS in order to trigger a Function. Eventing Controller also dispatches messages to subscribers such as a Function or another workload.

## Event types

Eventing supports both Cloud Events and legacy events. Event Publisher Proxy converts legacy events to Cloud Events and adds the `sap.kyma.custom` prefix.

For a Subscription Custom Resource, the fully qualified event type takes the sample form of `sap.kyma.custom.commerce.order.created.v1` or `sap.kyma.custom.commerce.Account.Root.Created.v1`.

The event type is composed of the following components:
- Prefix: `sap.kyma.custom`
- Application: `commerce`
- Event: can have two or more segments separated by `.`; for example, `order.created` or `Account.Root.Created`
- Version: `v1`

For publishers, the event type takes this sample form:
- `order.created` or `Account.Root.Created` for legacy events coming from the `commerce` application
- `sap.kyma.custom.commerce.order.created.v1` or `sap.kyma.custom.commerce.AccountRoot.Created.v1` for Cloud Events

>**NOTE:** If the event contains more than two segments, Eventing combines them into two segments when creating the underlying Eventing infrastructure. For example, `Account.Root.Created` becomes `AccountRoot.Created`.