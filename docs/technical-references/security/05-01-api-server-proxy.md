---
title: API Server Proxy chart
type: Configuration
---

To configure the API Server Proxy chart, override the default values of its `values.yaml` file. This document describes parameters that you can configure.

>**TIP:** To learn more about how to use overrides in Kyma, see the following documents:
>* [Helm overrides for Kyma installation](/root/kyma/#configuration-helm-overrides-for-kyma-installation)
>* [Top-level charts overrides](/root/kyma/#configuration-helm-overrides-for-kyma-installation-top-level-charts-overrides)

## Configurable parameters

This table lists the configurable parameters, their descriptions, and default values:

| Parameter | Description | Default value |
|-----------|-------------|---------------|
| **port.secure** | Specifies the port that exposes API Server Proxy through the load balancer. | `9443` |
| **port.insecure** | Specifies the port that exposes API Server Proxy through Istio Ingress. | `8444` |
| **hpa.minReplicas** | Defines the initial number of created API Server Proxy instances. | `1` |
| **hpa.maxReplicas** | Defines the maximum number of created API Server Proxy instances. | `3` |
| **hpa.metrics.resource.targetAverageUtilization** | Specifies the average percentage of a given instance memory utilization. After exceeding this limit, Kubernetes creates another API Server Proxy instance. | `50` |
