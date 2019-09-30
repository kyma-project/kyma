---
title: Asset Upload Service
type: Matrics
---

This table shows the Asset Upload Service custom metrics, their types, and descriptions.

| Name | Type | Decription |
|------|-------------|------|
| `assetstore_upload_service_http_request_duration_seconds` | histogram | Specifies a number of HTTP requests the service processes in a given time series. |
| `assetstore_upload_service_http_request_returned_status_code` | counter | Specifies a number of different HTTP response status codes in a given time series. |

Apart from the custom metrics, the Asset Upload Service also exposes default Prometheus metrics for [Go applications](https://prometheus.io/docs/guides/go-application/).

See the [Monitoring](/components/monitoring) documentation to learn more about monitoring and metrics in Kyma.
