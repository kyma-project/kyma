---
title: Architecture
---

Kyma Eventing uses NATS to implement an Event Publisher Proxy and Eventing Controller.

This diagram shows how the Eventing components work together. //TODO proper diagram - this is a placeholder

![Eventing implementation](./assets/eventing-implementation.svg)

## Event Publisher Proxy

The Event Publisher Proxy component receives legacy and Cloud Event publishing requests from the cluster workloads (microservice or Serverless Functions) and redirects them to the Enterprise Messaging Service Cloud Event Gateway. It also fetches a list of subscriptions for a connected application.

## Eventing Controller

The Eventing Controller component manages the internal infrastructure in order to receive an event. The Controller watches subscription Custom Resource Definitions. When an event is received in an Application, it lays down the Eventing infrastructure in NATS in order to trigger a Function.

## Event types

Kyma Eventing supports both cloud events and legacy events. Kyma converts legacy events to cloud events and adds the prefix `sap.kyma`.

For a Subscription Custom Resource, the fully qualified event type is in the form `sap.kyma.custom.commerce.order.created.v1`.
​
For publishers, it is:
- `order.created` for legacy-events coming from the `commerce` application
- `sap.kyma.custom.commerce.order.cre`
