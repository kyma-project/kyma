---
title: Istio chart
---

To configure the Istio chart, override the default values of its [`values.yaml`](https://github.com/kyma-project/kyma/blob/main/resources/istio-configuratio/nvalues.yaml) file. This document describes parameters that you can configure.

The Istio installation in Kyma uses the [IstioOperator](https://istio.io/docs/reference/config/istio.operator.v1alpha1/) API. Kyma provides default IstioOperator configurations for the production and evaluation profiles, You can add a custom IstioOperator definition that overrides the default settings. See the default `values.yaml` files for the Istio [evaluation](https://github.com/kyma-project/kyma/blob/main/resources/istio-configuration/profile-evaluation.yaml) and [production](https://github.com/kyma-project/kyma/blob/main/resources/istio-configuration/profile-production.yaml) profiles.

>**TIP:** See how to [change Kyma settings](../../04-operation-guides/operations/03-change-kyma-config-values.md).

## Configurable parameters

This table lists the configurable parameters, their descriptions, and default values:

| Parameter |  Description | Default value |
|-------|-------|:--------:|
| **helmValues.lobal.proxy.resources.requests.cpu** | Defines requests for CPU resources for the Proxy component. | `75m` |
| **helmValues.global.proxy.resources.requests.memory** | Defines requests for memory resources for the Proxy component. | `64Mi` |
| **helmValues.global.proxy.resources.limits.cpu** | Defines limits for CPU resources for the Proxy component. | `250m` |
| **helmValues.global.proxy.resources.limits.memory** | Defines limits for memory resources for the Proxy component. | `256Mi` |
| **components.ingressGateways.config.hpaSpec.maxReplicas** | Defines the maximum number of replicas managed by the autoscaler. | `5` |
| **components.ingressGateways.config.hpaSpec.minReplicas** | Defines the minimum number of replicas managed by the autoscaler. | `1` |
| **components.ingressGateways.resources.limits.cpu** | Defines limits for CPU resources for the Ingress Gateway component. | `200m` |
| **components.ingressGateways.resources.limits.memory** | Defines limits for memory resources for the Ingress Gateway component. | `1024Mi` |
| **components.ingressGateways.resources.requests.cpu** | Defines requests for CPU resources for the Ingress Gateway component. | `100m` |
| **components.ingressGateways.resources.requests.memory** | Defines requests for memory resources for the Ingress Gateway component.| `128Mi` |
| **components.pilot.resources.limits.cpu** | Defines limits for CPU resources for the Pilot component. | `500m` |
| **components.pilot.resources.limits.memory** | Defines limits for memory resources for the Pilot component. | `1024Mi` |
| **components.pilot.resources.requests.cpu** | Defines requests for CPU resources for the Pilot component. | `250m` |
| **components.pilot.resources.requests.memory** | Defines requests for memory resourcesfor the Pilot component. | `512Mi` |
