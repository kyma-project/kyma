---
title: Overview
---

Eventing in Kyma allows you to easily integrate external applications. Under the hood, Eventing implements [NATS](https://nats.io) to ensure Kyma receives business events from external sources and is able to trigger business flows using Functions or services. NATS provides an abstraction layer where data is encoded and framed as a message and sent by a publisher. The message is received, decoded, and processed by one or more subscribers.

To learn more about how to use Eventing, try the [Trigger a Function with an event](/components/serverless/#tutorials-trigger-a-function-with-an-event) and [Trigger the microservice with an event](/root/getting-started/#getting-started-trigger-the-microservice-with-an-event)  tutorials.