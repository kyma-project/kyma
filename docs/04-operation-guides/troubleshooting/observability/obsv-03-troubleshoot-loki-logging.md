---
title: Loki doesn't show the logs you want to see
---

## Condition

Loki shows fewer logs than you would like to see.

## Cause

There's a fixed logs retention time and size. If the default time is exceeded, the oldest logs are removed first.

## Remedy

- If you want to see logs older than 5 days, [increase the retention period](../../operations/obsv-02-adjust-loki.md#adjust-log-retention-period).
- If your logs aren't saved because your volume size is too small, [expand the Persistent Volume Claims](../../operations/obsv-02-adjust-loki.md#adjust-volume-size).
- If your logs persistently exceed the ingestion limit, [expand the ingestion rate](../../operations/obsv-02-adjust-loki.md#adjust-ingestion-limit).
