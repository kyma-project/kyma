---
title: Catalog sub-chart
type: Configuration
---

To configure the Catalog sub-chart, override the default values of its `values.yaml` file. This document describes parameters that you can configure.

>**TIP:** To learn more about how to use overrides in Kyma, see the following documents:
>* [Helm overrides for Kyma installation](/root/kyma/#configuration-helm-overrides-for-kyma-installation)
>* [Sub-charts overrides](/root/kyma/#configuration-helm-overrides-for-kyma-installation-sub-chart-overrides)

## Configurable parameters

This table lists the configurable parameters, their descriptions, and default values:

| Parameter | Description | Default value |
|-----------|-------------|---------------|
| **controllerManager.resources.requests.cpu** | Defines requests for CPU resources. | `100m` |
| **controllerManager.resources.requests.memory** | Defines requests for memory resources. | `20Mi` |
| **controllerManager.resources.limits.cpu** | Defines limits for CPU resources. | `100m` |
| **controllerManager.resources.limits.memory** | Defines limits for memory resources. | `30Mi` |
| **controllerManager.verbosity** | Defines log severity level. The possible values range from 0-10. | `10` |
| **webhook.resources.requests.cpu** | Defines requests for CPU resources. | `100m` |
| **webhook.resources.requests.memory** | Defines requests for memory resources. | `20Mi` |
| **webhook.resources.limits.cpu** | Defines limits for CPU resources. | `100m` |
| **webhook.resources.limits.memory** | Defines limits for memory resources. | `30Mi` |
| **webhook.verbosity** | Defines log severity level. The possible values range from 0-10. | `10` |
| **controllerManager.brokerRelistIntervalActivated** | Specifies whether or not the controller supports a `--broker-relist-interval` flag. If this is set to `true`, brokerRelistInterval will be used as the value for that flag. | `true` |
| **controllerManager.brokerRelistInterval** | Specifies how often the controller relists the catalogs of ready brokers. The duration format is 20m, 1h, etc. | `24h` |
| **originatingIdentityEnabled** | Enables the OriginatingIdentity feature which controls whether the controller includes originating identity in the header of requests sent to brokers and persisted under a CustomResource. | `true` |
| **asyncBindingOperationsEnabled** | Enables support for asynchronous binding operations. | `true` |
| **namespacedServiceBrokerDisabled** | Disables support for Namespace-scoped brokers. | `false` |
| **securityContext** | Gives the opportunity to run container as non-root by setting a securityContext. For example: `securityContext: { runAsUser: 1001 }` | `{}` |
