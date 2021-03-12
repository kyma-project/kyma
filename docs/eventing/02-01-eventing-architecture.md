---
title: Architecture
---

Kyma Eventing uses NATS to implement an Event Publisher Proxy and Eventing Controller which work together to process and deliver events in Kyma.

## Event Publisher Proxy

The Event Publisher Proxy component receives legacy and Cloud Event publishing requests from the cluster workloads (microservice or Serverless Functions) and redirects them to the NATS server. It also fetches a list of subscriptions for a connected application.

## Eventing Controller

The Eventing Controller component manages the internal infrastructure in order to receive an event. The Controller watches subscription Custom Resource Definitions. When an event is received in an Application, it lays down the Eventing infrastructure in NATS in order to trigger a Function. The Eventing Controller also dispatches messages to subscribers such as a Serverless Function or another workload.

## Event types

Kyma Eventing supports both cloud events and legacy events. Kyma converts legacy events to cloud events and adds the prefix `sap.kyma.custom`.

For a Subscription Custom Resource, the fully qualified event type is in the form `sap.kyma.custom.commerce.order.created.v1`.
​
For publishers, it is:
- `order.created` for legacy-events coming from the `commerce` application
- `sap.kyma.custom.commerce.order.created.v1` for CloudEvents.
