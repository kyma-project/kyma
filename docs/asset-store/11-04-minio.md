---
title: Minio
type: Metrics
---

As an external, open-source file storage solution, Minio exposes its own metrics. See the [official documentation](https://github.com/minio/minio/tree/master/docs/metrics) for details. The Asset Stores comes with a preconfigured ServiceMonitor CR that enables Prometheus to scrap Minio metrics. Using the metrics, you can create your own Grafana dashboard or reuse the dashboard that is already prepared.
