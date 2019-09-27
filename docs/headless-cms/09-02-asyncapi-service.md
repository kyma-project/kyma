---
title: CMS AsyncAPI Service
type: Metrics
---

This table shows the CMS AsyncAPI Service custom matrics, their descriptions, and types.

| Name | Type | Description |
|------|-------------|------|
| `cms_services_http_request_and_mutation_duration_seconds` | histogram | Specifies a number of assets that the service received for processing and mutated within a given time series. |
| `cms_services_http_request_and_validation_duration_seconds` | histogram | Specifies a number of assets that the service received for processing and validated within a given time series. |
| `cms_services_handle_mutation_status_code` | counter | Specifies a number of `200` and `304` HTTP response status codes in a given time series returned by the mutation handler. |
| `cms_services_handle_mutation_status_code` | counter | Specifies a number of `200` and `422` HTTP response status codes in a given time series returned by the validation handler. |
