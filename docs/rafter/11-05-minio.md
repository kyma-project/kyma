---
title: MinIO
type: Metrics
---

As an external, open-source file storage solution, MinIO exposes its own metrics. See the [official documentation](https://github.com/minio/minio/tree/master/docs/metrics) for details. Rafter comes with a preconfigured ServiceMonitor CR that enables Prometheus to scrap MinIO metrics. Using the metrics, you can create your own Grafana dashboard or reuse the dashboard that is already prepared.

Apart from the custom metrics, MinIO also exposes default Prometheus metrics for [Go applications](https://prometheus.io/docs/guides/go-application/).

To see a complete list of metrics, run this command:

```bash
kubectl -n kyma-system port-forward svc/rafter-minio 9000
```

To check the metrics, open a new terminal window and run:

```bash
curl http://localhost:9000/minio/prometheus/metrics
```

> **TIP:** To use these commands, you must have a running Kyma cluster and kubectl installed. If you cannot access port `9000`, redirect the metrics to another one. For example, run `kubectl -n kyma-system port-forward svc/rafter-minio 3000:9000` and update the port in the localhost address.

See the [Monitoring](/components/monitoring) documentation to learn more about monitoring and metrics in Kyma.
