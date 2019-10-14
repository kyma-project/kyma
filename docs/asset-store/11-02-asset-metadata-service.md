---
title: Asset Metadata Service
type: Metrics
---

This table shows the Asset Metadata Service custom metrics, their types, and descriptions.

| Name | Type | Description |
|------|-------------|------|
| `assetstore_metadata_service_http_request_duration_seconds` | histogram | Specifies a number of HTTP requests the service processes in a given time series. |
| `assetstore_metadata_service_http_request_returned_status_code` | counter | Specifies a number of different HTTP response status codes in a given time series. |

Apart from the custom metrics, the Asset Metadata Service also exposes default Prometheus metrics for [Go applications](https://prometheus.io/docs/guides/go-application/).

To see a complete list of the custom and Go metrics, run the following command:

```bash
kubectl -n kyma-system port-forward svc/assetstore-asset-metadata-service 80
```

To check the metrics, open a new terminal window and run:

```bash
curl http://localhost:80/metrics
```

> **TIP:** Before you use the command, make sure you have a running Kyma cluster and kubectl installed. If you cannot access port 80, redirect the metrics to another one. For example, run: `kubectl -n kyma-system port-forward svc/assetstore-asset-metadata-service 8080:80:` and update the port in the localhost address.

See the [Monitoring](/components/monitoring) documentation to learn more about monitoring and metrics in Kyma.
