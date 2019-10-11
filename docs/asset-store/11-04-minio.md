---
title: Minio
type: Metrics
---

As an external, open-source file storage solution, Minio exposes its own metrics. See the [official documentation](https://github.com/minio/minio/tree/master/docs/metrics) for details. The Asset Stores comes with a preconfigured ServiceMonitor CR that enables Prometheus to scrap Minio metrics. Using the metrics, you can create your own Grafana dashboard or reuse the dashboard that is already prepared.

Apart from the custom metrics, Minio also exposes default Prometheus metrics for [Go applications](https://prometheus.io/docs/guides/go-application/).

To see a complete list of the Go metrics, run the following command:

```bash
kubectl -n kyma-system port-forward svc/assetstore-minio 9000
```

Now open a browser and access [http://localhost:9000/minio/prometheus/metrics](http://localhost:9000/minio/prometheus/metrics) to check the metrics.

> **TIP:** Before you use the command, make sure you have a running Kyma cluster and kubectl installed. If you cannot access the 9000 port, redirect the metrics to another one. For example, run: `kubectl -n kyma-system port-forward svc/assetstore-minio 3000:9000` and update the port in the localhost address in your browser.

See the [Monitoring](/components/monitoring) documentation to learn more about monitoring and metrics in Kyma.
