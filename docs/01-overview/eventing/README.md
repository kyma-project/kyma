---
title: What is Eventing in Kyma?
---

## Overview

With Kyma Eventing, you can focus on your business workflows and trigger them with events to implement asynchronous flows within Kyma. Generally, eventing consists of event producers (or publishers) and consumers (or subscribers) that send events to or receive events from an event processing backend.

The objective of Eventing in Kyma is to simplify the process of publishing and subscribing to events. Kyma uses proven eventing backend technology to provide a seamless experience to the user with their end-to-end business flows. The user does not have to implement or integrate any intermediate backend or protocol.

Kyma Eventing uses the following technology:
- [NATS JetStream](https://docs.nats.io/) as backend within the cluster
- [HTTP POST](https://www.w3schools.com/tags/ref_httpmethods.asp) requests to simplify sending and receiving events
- Declarative [Subscription CR](../../05-technical-reference/00-custom-resources/evnt-01-subscription.md) to subscribe to events

## Kyma Eventing flow

Kyma Eventing follows the PubSub messaging pattern: Kyma publishes messages to a messaging backend, which filters these messages and sends them to interested subscribers. Kyma does not send messages directly to the subscribers as shown below:

![PubSub](./assets/pubsub.svg)

Eventing in Kyma from a user’s perspective works as follows:

- Offer an HTTP end point, for example a Function to receive the events.
- Specify the events the user is interested in using the Kyma [Subscription CR](../../05-technical-reference/00-custom-resources/evnt-01-subscription.md).
- Send [CloudEvents](https://cloudevents.io/) or legacy events (deprecated) to the following HTTP end points on our [Event Publishing Proxy](https://github.com/kyma-project/kyma/tree/main/components/event-publisher-proxy) service.
    - `/publish` for CloudEvents.
    - `<application_name>/v1/events` for legacy events.

For more information, read the [Eventing architecture](../../05-technical-reference/00-architecture/evnt-01-architecture.md).