---
title: Loki doesn't show the logs you want to see
---

## Condition

Loki shows fewer logs than you would like to see.

## Cause

By default, Loki stores up to 30 GB of data for a maximum of 5 days, with maximum ingestion rate of 3 MB/s. If the default size or time is exceeded, the oldest logs are removed first.

## Remedy

- If you want to see logs older than 5 days, [increase the retention period](../../04-operation-guides/operations/obsv-02-adjust-loki.md#adjust-log-retention-period).
- If your logs aren't saved because your volume size is too small, [expand the Persistent Volume Claims](../../04-operation-guides/operations/obsv-02-adjust-loki.md#adjust-volume-size).
- If your logs persistently exceed the ingestion limit, [expand the ingestion rate](../../04-operation-guides/operations/obsv-02-adjust-loki.md#adjust-ingestion-limit).