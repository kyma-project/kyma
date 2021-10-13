---
title: Blocked services in API Gateway Controller
---

API Gateway Controller uses a blocklist of services for which it does not create either a Virtual Service or Oathkeeper Access Rules. As a result, these services cannot be exposed. Every time a user creates a new APIRule custom resource (CR) for a service, API Gateway Controller checks the name of the service specified in the CR against the blocklist. If the name of the service matches a blocklisted entry, API Gateway Controller sets an appropriate validation status on the APIRule CR created for that service.

>**TIP:** For more information, read about the [API CR statuses](./00-custom-resources/apix-01-apirule.md#status-codes).

The blocklist works as a security measure and prevents users from exposing vital internal services of Kubernetes and Istio.
