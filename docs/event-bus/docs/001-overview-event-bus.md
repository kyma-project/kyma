---
title: Overview
---

Kyma Event Bus enables the integration of various external solutions with Kyma. The integration is achieved using the `publish-subscribe` messaging pattern that allows Kyma to receive business Events from different solutions, enrich them, and trigger business flows using lambdas or services defined in Kyma.

To learn how to write an HTTP service or lambda in Kyma, and handle the Event Bus published Events, check the **Services Programming Model** document.

> **NOTE:** The Event Bus is based on the [NATS Streaming](https://github.com/nats-io/nats-streaming-server/releases) open source log-based streaming system for cloud-native applications, which is a brokered messaging middleware. The Event Bus provides **at-least-once** delivery guarantees.
