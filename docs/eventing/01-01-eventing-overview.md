---
title: Overview
---

Eventing in Kyma allows you to easily integrate external applications. Under the hood, Kyma Eventing implements [NATS](https://nats.io) to ensure Kyma receives business events from external sources and is able to trigger business flows using Functions or services. NATS provides an abstraction layer where data is encoded and framed as a message and sent by a publisher. The message is received, decoded, and processed by one or more subscribers.