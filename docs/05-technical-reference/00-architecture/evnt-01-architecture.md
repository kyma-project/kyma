---
title: Eventing Architecture
---

The Event Publisher Proxy and the Eventing Controller are the two main components of Eventing. They work together to connect to the default NATS backend and process/deliver events in Kyma. See the [Event processing and delivery](../evnt-01-event-processing.md) document for more information.

## Event Publisher Proxy

The Event Publisher Proxy component receives legacy and Cloud Event publishing requests from the cluster workloads (microservices or Functions). It converts any legacy events to Cloud Events. Then, it redirects events to the NATS server. It also fetches a list of subscriptions for a connected application.

## Eventing Controller

The Eventing Controller component manages the internal infrastructure in order to receive an event. The Controller watches Subscription Custom Resource Definitions. When an event is received in an Application, it lays down the Eventing infrastructure in NATS in order to trigger a Function. The Eventing Controller also dispatches messages to subscribers such as a Function or another workload.

## Event types

Eventing supports both Cloud Events and legacy events. The Event Publisher Proxy converts legacy events to Cloud Events and adds the `sap.kyma.custom` prefix.

For a Subscription Custom Resource, the fully qualified event type takes the sample form of `sap.kyma.custom.commerce.order.created.v1`. The event type is composed of the following components:

- Prefix: `sap.kyma.custom`
- Application: `commerce`
- Event: `order.created`
- Version: `v1`
​
For publishers, the event type takes this sample form:
- `order.created` for legacy events coming from the `commerce` application
- `sap.kyma.custom.commerce.order.created.v1` for Cloud Events
