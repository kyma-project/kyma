---
title: Allowed domains in the API Gateway Controller
type: Details
---

You can restrict the set of domains for which the API Gateway Controller allows to expose services. When the feature is enabled, every time a user creates a new APIRule custom resource (CR) for a service, the API Gateway Controller checks the domain of the service specified in the CR against the allowlist. If the domain of the service matches an allowed entry, the API Gateway Controller creates a Virtual Service and Oathkeeper Access Rules for the service according to the details specified in the CR. If the domain is not allowed, the Controller creates neither a Virtual Service nor Oathkeeper Access Rules and, as a result, does not expose the service.

If the domain does not match the allowlist, the API Gateway Controller sets an appropriate validation status on the APIRule CR created for that service.

>**TIP:** For more information, read about the [Api CR statuses](#custom-resource-api-rule-status-codes).

By default, the feature is disabled and all domains are allowed

To enable the allowlist mechanism, override the value of the **config.enableDomainAllowList** parameter in the API Gateway chart.

>**TIP:** To learn more about how to use overrides in Kyma, see the following documents:
>* [Helm overrides for Kyma installation](/root/kyma/#configuration-helm-overrides-for-kyma-installation)
>* [Top-level charts overrides](/root/kyma/#configuration-helm-overrides-for-kyma-installation-top-level-charts-overrides)
