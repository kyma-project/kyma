---
title: What is Eventing in Kyma?
---

Eventing in Kyma is an area that:

- enables users to use events to implement asynchronous flows within Kyma.
- is driven by [NATS JetStream](https://docs.nats.io/) backend within the cluster.
- allows the user to focus only on the business workflow and trigger them using events.
- simplifies sending and receiving events using HTTP POST requests.
- uses declarative [Subscription CR](../../../05-technical-reference/00-custom-resources/evnt-01-subscription.md) to subscribe to events.