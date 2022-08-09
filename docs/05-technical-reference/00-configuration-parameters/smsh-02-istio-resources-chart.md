---
title: Istio Resources chart
---

To configure the Istio Resources chart, override the default values of its [`values.yaml`](https://github.com/kyma-project/kyma/blob/main/resources/istio-resources/values.yaml) file. This document describes parameters that you can configure.

>**TIP:** See how to [change Kyma settings](../../04-operation-guides/operations/03-change-kyma-config-values.md).

## Configurable parameters

This table lists the configurable parameters, their descriptions, and default values:

| Parameter | Description | Default value |
|-----------|-------------|---------------|
| **monitoring.enabled** | Allows enabling Istio monitoring options. | `true` |
| **monitoring.dashboards.enabled** | Enables Istio monitoring dashboards. Requires **monitoring.enabled** set to `true`.| `true` |
| **monitoring.istio.Service.Monitor.enabled** | Enables Istio ServiceMonitor. Requires **monitoring.enabled** set to `true`. | `true` |
