---
title: Logging
---

## Overview

For in-cluster logging, Kyma uses [Loki](https://github.com/grafana/loki), a Prometheus-like log management system.

As a beta feature, Kyma ships the telemetry component, which supports providing your own output configuration for Fluent Bit. With this, you can integrate your own observability systems outside the Kyma cluster with the Kyma backend. If you want to use this feature, [enable the telemetry component](../../../04-operation-guides/operations/obsv-00-enable-configurable-logging.md).

## Limitations

In the production profile, Loki stores up to **30 GB** of data for a maximum of **5 days**, with maximum ingestion rate of 3 MB/s. If the default time is exceeded, the oldest logs are removed first.

The evaluation profile has lower limits. For more information about profiles, see [Install Kyma: Choose resource consumption](../../../04-operation-guides/operations/02-install-kyma.md#choose-resource-consumption).
