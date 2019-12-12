---
title: Whitelisted domains in the API Gateway Controller
type: Details
---

The API Gateway Controller uses a whitelist of domains for which it allows to expose services. Every time a user creates a new APIRule custom resource (CR) for a service, the API Gateway Controller checks the domain of the service specified in the CR against the whitelist. If the domain of the service matches a whitelisted entry, the API Gateway Controller creates a Virtual Service and Oathkeeper Access Rules for the service according to the details specified in the CR. If the domain is not whitelisted, the Controller creates neither a Virtual Service nor Oathkeeper Access Rules and, as a result, does not expose the service.

If the domain does not match the whitelist, the API Gateway Controller sets an appropriate validation status on the APIRule CR created for that service.

>**TIP:** Read [this](#custom-resource-api-rule-status-codes) document to learn more about the Api CR statuses.

By default, the only whitelisted domain is the domain of the Kyma cluster.
