---
title: CMS AsyncAPI Service
type: Metrics
---

This table shows the CMS AsyncAPI Service custom metrics, their types, and descriptions.

| Name | Type | Description |
|------|-------------|------|
| `cms_services_http_request_and_mutation_duration_seconds` | histogram | Specifies a number of assets that the service received for processing and mutated within a given time series. |
| `cms_services_http_request_and_validation_duration_seconds` | histogram | Specifies a number of assets that the service received for processing and validated within a given time series. |
| `cms_services_handle_mutation_status_code` | counter | Specifies a number of different HTTP response status codes in a given time series. |
| `cms_services_handle_mutation_status_code` | counter | Specifies a number of different HTTP response status codes in a given time series. |

Apart from the custom metrics, the CMS AsyncAPI Service also exposes default Prometheus metrics for [Go applications](https://prometheus.io/docs/guides/go-application/).

To see a complete list of the custom and Go metrics, run the following command:

```bash
kubectl -n kyma-system port-forward svc/cms-cms-asyncapi-service 80
```

Now open a browser and access [http://localhost:80/metrics](http://localhost:80/metrics) to check the metrics.

> **TIP:** Before you use the command, make sure you have a running Kyma cluster and kubectl installed. If you cannot access the 8080 port, redirect the metrics to another one. For example, run: `kubectl -n kyma-system port-forward svc/cms-cms-asyncapi-service 8080:80` and update the port in the localhost address in your browser.

See the [Monitoring](/components/monitoring) documentation to learn more about monitoring and metrics in Kyma.
