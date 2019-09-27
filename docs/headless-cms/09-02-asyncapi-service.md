---
title: CMS AsyncAPI Service
type: Metrics
---

This table shows the CMS AsyncAPI Service custom matrics, their types, and descriptions.

| Name | Type | Description |
|------|-------------|------|
| `cms_services_http_request_and_mutation_duration_seconds` | histogram | Specifies a number of assets that the service received for processing and mutated within a given time series. |
| `cms_services_http_request_and_validation_duration_seconds` | histogram | Specifies a number of assets that the service received for processing and validated within a given time series. |
| `cms_services_handle_mutation_status_code` | counter | Specifies a number of different HTTP response status codes in a given time series. |
| `cms_services_handle_mutation_status_code` | counter | Specifies a number of different HTTP response status codes in a given time series. |

Apart from the custom metrics, the CMS AsyncAPI Service also exposes:

- default metrics instrumented by [kubebuilder](https://book.kubebuilder.io/)
- default Prometheus metrics for [Go applications](https://prometheus.io/docs/guides/go-application/#how-go-exposition-works)

See the [Monitoring](/components/monitoring) documentation to learn more about monitoring and metrics in Kyma.
