---
title: Architecture
---

This document provides an overview of the logging architecture in Kyma. It describes the information sources from which promtail extract logs to feed to Loki.

## Agent (Promtail)
Agent is responsible for obtaining reliable metadata that is consistent with the metadata associated with the time series / metrics. To achieve this agent use the same service discovery and label relabelling libraries as Prometheus. Agent packed up in a daemon that discovers targets, produces metadata labels and tails log files to produce stream of logs, which will be buffered on client side and the sent to the service.

#### Life of a Write Request
The server-side components on the write path wii mirror the [Cortex](https://github.com/cortexproject/cortex) architecture.
* Writes will first hit the Distributor, which is responsible for distributing and replacing the writes to the ingesters. Loki use the Cortex consistent hash ring and distribute writes based on a hash of the entire metadata.
* Next writes will hit a 'log ingester' which batches up writes for the same stream in memory in to 'log chunks'. When chunks reach a predefined size or age, periodically flushed to the Cortex chunk store.
* The Cortex chunk store will be updated to reduce copying of chunk data on the read and write path and add support for writing chunks of Grafana.

#### Log Chunks
A chunk is all logs for a given label set over a certain period. the chunks support appends, seek and streaming reads.

#### Life of a Query Request
As chunks are larger than Prometheus/Cortex chunks (Cortex chunks are max 1KB in size), it is not possible to load and decompress them in their entirety. Loki support streaming and iterating over them, only decompressing the parts necessary 


Further information consult the original [design doc](https://docs.google.com/document/d/11tjK_lvp1-SVsFZjgOTr1vV3-q6vBAsZYIQ5ZeYBkyM/view)
