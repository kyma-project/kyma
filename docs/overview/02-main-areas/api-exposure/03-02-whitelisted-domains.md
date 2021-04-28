---
title: Allowed domains in the API Gateway Controller
type: Details
---

The API Gateway Controller uses an allowlist of domains for which it allows to expose services. Every time a user creates a new APIRule custom resource (CR) for a service, the API Gateway Controller checks the domain of the service specified in the CR against the allowlist. If the domain of the service matches an allowed entry, the API Gateway Controller creates a Virtual Service and Oathkeeper Access Rules for the service according to the details specified in the CR. If the domain is not allowed, the Controller creates neither a Virtual Service nor Oathkeeper Access Rules and, as a result, does not expose the service.

If the domain does not match the allowlist, the API Gateway Controller sets an appropriate validation status on the APIRule CR created for that service.

>**TIP:** For more information, read about the [Api CR statuses](#custom-resource-api-rule-status-codes).

By default, the only allowed domain is the domain of the Kyma cluster.
