---
title: Architecture
---

This document provides an overview of the logging architecture in Kyma. It describes the information sources from which promtail extract logs to feed to Loki.

## Agent (Promtail)
Promtail is the agent responsible for collecting reliable metadata, consistent with the time series, or metrics metadata. To achieve this, the agent uses the same service discovery and label relabelling libraries as Prometheus. The agent is wrapped in a daemon that discovers targets, produces metadata labels, and tails log files to produce a stream of logs buffered on the client side and then sent to the service.

## Life of a write request
The server-side components on the write path will mirror the [Cortex](https://github.com/cortexproject/cortex) architecture.
* Writes will first hit the Distributor, which is responsible for distributing and replacing the writes to the ingesters. Loki use the Cortex consistent hash ring and distribute writes based on a hash of the entire metadata.
* Next writes will hit a 'log ingester' which batches up writes for the same stream in memory in to 'log chunks'. When chunks reach a predefined size or age, periodically flushed to the Cortex chunk store.
* The Cortex chunk store will be updated to reduce copying of chunk data on the read and write path and add support for writing chunks of google cloud storage.

#### Log Chunks
A log chunk consists of all logs for metadata, such as labels, collected over a certain time period. Log chunks support append, seek, and stream operations on requests.

#### Life of a Query Request
As chunks are larger than Prometheus/Cortex chunks (Cortex chunks are max 1KB in size), it is not possible to load and decompress them in their entirety. 
To solve this problem Loki support streaming and iterating over them, therefor loki will decompress only necessary chunk parts.


For further information, see the [design document](https://docs.google.com/document/d/11tjK_lvp1-SVsFZjgOTr1vV3-q6vBAsZYIQ5ZeYBkyM/view).
