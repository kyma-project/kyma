---
title: Architecture
---

Eventing uses NATS to implement an Event Publisher Proxy and Eventing Controller which work together to process and deliver events in Kyma.

## Event Publisher Proxy

The Event Publisher Proxy component receives legacy and cloud event publishing requests from the cluster workloads (microservices or Functions) and redirects them to the NATS server. It also fetches a list of subscriptions for a connected application.

## Eventing Controller

The Eventing Controller component manages the internal infrastructure in order to receive an event. The Controller watches Subscription Custom Resource Definitions. When an event is received in an Application, it lays down the Eventing infrastructure in NATS in order to trigger a Function. The Eventing Controller also dispatches messages to subscribers such as a Function or another workload.

## Event types

Kyma Eventing supports both cloud events and legacy events. Kyma converts legacy events to cloud events and adds the `sap.kyma.custom` prefix.

For a Subscription Custom Resource, the fully qualified event type takes the form of `sap.kyma.custom.commerce.order.created.v1`. The event type is composed of the following components:

- Prefix: `sap.kyma.custom`
- Application: `commerce`
- Event: `order.created`
- Version: `v1`
​
For publishers, the event type takes this form:
- `order.created` for legacy events coming from the `commerce` application
- `sap.kyma.custom.commerce.order.created.v1` for cloud events.
