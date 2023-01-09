---
title: Tracing
---

## Overview

For in-cluster tracing, Kyma uses [Jaeger](https://github.com/jaegertracing). With this distributed tracing system, you can analyze the path of a request chain going through your distributed applications. This information helps you to, for example, troubleshoot your applications, or optimize the latency and performance of your solution.

Kyma's [telemetry component](./../telemetry/README.md) supports providing your own output configuration for traces. With this, you can connect your own observability systems inside or outside the Kyma cluster.

## Limitations

In the production profile, Jaeger has no persistence enabled and keeps up to 10.000 traces stored in-memory. The oldest records are removed first. The evaluation profile has lower limits. For more information about profiles, see [Install Kyma: Choose resource consumption](../../../04-operation-guides/operations/02-install-kyma.md#choose-resource-consumption).
