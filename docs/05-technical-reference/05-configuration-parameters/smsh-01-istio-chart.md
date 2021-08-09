---
title: Istio <!--Operator--> chart
---

To configure the Istio chart and, override the default values of its [`values.yaml`](https://github.com/kyma-project/kyma/blob/main/resources/istio-configuratio/nvalues.yaml) file. This document describes parameters that you can configure.

The Istio installation in Kyma uses the [IstioOperator](https://istio.io/docs/reference/config/istio.operator.v1alpha1/) API. Kyma provides default IstioOperator configurations for the production and evaluation profiles, You can add a custom IstioOperator definition that overrides the default settings. See the default `values.yaml` files for the [evaluation](https://github.com/kyma-project/kyma/blob/main/resources/istio-configuration/profile-evaluation.yaml) and [production](https://github.com/kyma-project/kyma/blob/main/resources/istio-configuration/profile-production.yaml) profiles.

>**TIP:** See how to [change Kyma settings](../../04-operation-guides/operations/03-change-kyma-config-values.md).

## Configurable parameters

This table lists the configurable parameters, their descriptions, and default values:

| Parameter |  Description | Default value |
|-------|-------|:--------:|
| **global.proxy.resources.requests.cpu** | Defines requests for CPU resources. | 75m |
| **global.proxy.resources.requests.memory** | Defines requests for memory resources. | 64Mi |
| **global.proxy.resources.limits.cpu** | Defines limits for CPU resources. | 250m |
| **global.proxy.resources.limits.memory** | Defines limits for memory resources. | 256Mi |
| **components.ingressGateways.config.hpaSpec.maxReplicas** | Defines the maximum number of . | 5 |
| **components.ingressGateways.config.hpaSpec.minReplicas** | Defines the minimum number of . | 1 |
