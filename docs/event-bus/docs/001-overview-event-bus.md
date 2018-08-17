---
title: Overview
type: Overview
---

Kyma Event Bus enables the integration of various external solutions with Kyma. The integration is achieved using the `publish-subscribe` messaging pattern that allows Kyma to receive business Events from different solutions, enrich them, and trigger business flows using lambdas or services defined in Kyma.

To learn how to write an HTTP service or a Lambda in Kyma and handle the event bus published events, please check the Services Programming Model [guide](013-service-programming-model.md) and the Lambda Programming Model [guide](../../serverless/docs/035-programming-model.md).

> **NOTE:** The Event Bus is based on the [NATS Streaming](https://github.com/nats-io/nats-streaming-server/releases) open source log-based streaming system for cloud-native applications, which is a brokered messaging middleware. The Event Bus provides **at-least-once** delivery guarantees.
