---
title: Loki doesn't show the logs you want to see
---

> **NOTE:** Loki is [deprecated](https://kyma-project.io/blog/2022/11/2/loki-deprecation/) and is planned to be removed. If you want to install a custom Loki stack, take a look at [this tutorial](https://github.com/kyma-project/examples/tree/main/loki).

## Condition

Loki shows fewer logs than you would like to see.

## Cause

There's a fixed logs retention time and size. If the default time is exceeded, the oldest logs are removed first.

## Remedy

- If you want to see logs older than 5 days, [increase the retention period](../../operations/obsv-02-adjust-loki.md#adjust-log-retention-period).
- If your logs aren't saved because your volume size is too small, [expand the PersistentVolumeClaims](../../operations/obsv-02-adjust-loki.md#adjust-volume-size).
- If your logs persistently exceed the ingestion limit, [expand the ingestion rate](../../operations/obsv-02-adjust-loki.md#adjust-ingestion-limit).
