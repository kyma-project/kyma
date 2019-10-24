---
title: Asset Upload Service
type: Metrics
---

This table shows the Asset Upload Service custom metrics, their types, and descriptions.

| Name | Type | Decription |
|------|-------------|------|
| `assetstore_upload_service_http_request_duration_seconds` | histogram | Specifies a number of HTTP requests the service processes in a given time series. |
| `assetstore_upload_service_http_request_returned_status_code` | counter | Specifies a number of different HTTP response status codes in a given time series. |

Apart from the custom metrics, the Asset Upload Service also exposes default Prometheus metrics for [Go applications](https://prometheus.io/docs/guides/go-application/).

To see a complete list of metrics, run this command:

```bash
kubectl -n kyma-system port-forward svc/assetstore-asset-upload-service 3000
```

To check the metrics, open a new terminal window and run:

```bash
curl http://localhost:3000/metrics
```

> **TIP:** To use these commands, you must have a running Kyma cluster and kubectl installed. If you cannot access port `3000`, redirect the metrics to another one. For example, run `kubectl -n kyma-system port-forward svc/assetstore-asset-upload-service 8080:3000` and update the port in the localhost address.

See the [Monitoring](/components/monitoring) documentation to learn more about monitoring and metrics in Kyma.
