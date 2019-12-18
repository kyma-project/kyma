---
title: AsyncAPI Service
type: Metrics
---

This table shows the AsyncAPI Service custom metrics, their types, and descriptions.

| Name | Type | Description |
|------|-------------|------|
| `rafter_services_http_request_and_mutation_duration_seconds` | histogram | Specifies the number of assets that the service received for processing and mutated within a given time series. |
| `rafter_services_http_request_and_validation_duration_seconds` | histogram | Specifies the number of assets that the service received for processing and validated within a given time series. |
| `rafter_services_handle_mutation_status_code` | counter | Specifies a number of different HTTP response status codes in a given time series. |
| `rafter_services_handle_validation_status_code` | counter | Specifies a number of different HTTP response status codes in a given time series. |

Apart from the custom metrics, the AsyncAPI Service also exposes default Prometheus metrics for [Go applications](https://prometheus.io/docs/guides/go-application/).

To see a complete list of metrics, run this command:

```bash
kubectl -n kyma-system port-forward svc/rafter-asyncapi-service 80
```

To check the metrics, open a new terminal window and run:

```bash
curl http://localhost:80/metrics
```

>**TIP:** To use these commands, you must have a running Kyma cluster and kubectl installed. If you cannot access port `80`, redirect the metrics to another one. For example, run `kubectl -n kyma-system port-forward svc/rafter-asyncapi-service 8080:80` and update the port in the localhost address.

See the [Monitoring](/components/monitoring) documentation to learn more about monitoring and metrics in Kyma.
