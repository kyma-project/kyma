---
title: Overview
---

Eventing in Kyma is used to send and receive events from applications. For example, you can subscribe to events from an external application, and when the user performs an action there, you can trigger your Function or microservice. To subscribe to events, you need to use the Kyma [Subscription CRD](../../../05-technical-reference/06-custom-resources/evnt-01-subscription.md).

Kyma supports [Cloud Events](https://cloudevents.io/) - a common specification for describing event data - and legacy events. Legacy events are converted to Cloud Events by Kyma.
