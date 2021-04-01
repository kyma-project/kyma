---
title: Architecture
---

Eventing uses NATS to implement the Event Publisher Proxy and the Eventing Controller which work together to process and deliver events in Kyma.

## Event Publisher Proxy

The Event Publisher Proxy component receives legacy and Cloud Event publishing requests from the cluster workloads (microservices or Functions) and redirects them to the NATS server. It also fetches a list of subscriptions for a connected application.

## Eventing Controller

The Eventing Controller component manages the internal infrastructure in order to receive an event. The Controller watches Subscription Custom Resource Definitions. When an event is received in an Application, it lays down the Eventing infrastructure in NATS in order to trigger a Function. The Eventing Controller also dispatches messages to subscribers such as a Function or another workload.

## Event types

Eventing supports both Cloud Events and legacy events. Kyma converts legacy events to Cloud Events and adds the `sap.kyma.custom` prefix.

For a Subscription Custom Resource, the fully qualified event type takes the form of `sap.kyma.custom.commerce.order.created.v1` or `sap.kyma.custom.commerce.Account.Root.Created.v1`. The event type is composed of the following components:

- Prefix: `sap.kyma.custom`
- Application: `commerce`
- Event: can have two or more segments separated by `.` (For example, `order.created` or `Account.Root.Created`)
- Version: `v1`

For publishers, the event type takes this form:
- `order.created` or `Account.Root.Created` for legacy events coming from the `commerce` application 
- `sap.kyma.custom.commerce.order.created.v1` or `sap.kyma.custom.commerce.AccountRoot.Created.v1` for Cloud Events.

>**NOTE:** In case the event contains more than two segments the underlying Eventing infrastructure will combine them into two segments (For example, `Account.Root.Created` will become `AccountRoot.Created`).
