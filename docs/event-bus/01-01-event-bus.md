---
title: Overview
---

Kyma Event Bus allows you to integrate various external solutions with Kyma. To achieve successful integration, the Event Bus uses the `publish-subscribe` messaging pattern that allows Kyma to receive business Events from different solutions, enrich them, and trigger business flows using lambdas or services defined in Kyma.

The Event Bus is based on the [NATS Streaming](https://github.com/nats-io/nats-streaming-server/releases) open-source log-based streaming system for cloud-native applications, which is a brokered messaging middleware. The Event Bus provides an **at-least-once** delivery guarantee meaning the messages are retransmitted to assure they are successfulluy delivered at least once.
