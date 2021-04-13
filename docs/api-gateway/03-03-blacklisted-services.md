---
title: Blocked services in the API Gateway Controller
type: Details
---

The API Gateway Controller uses a blocklist of services for which it does not create either a Virtual Service or Oathkeeper Access Rules. As a result, these services cannot be exposed. Every time a user creates a new APIRule custom resource (CR) for a service, the API Gateway Controller checks the name of the service specified in the CR against the blocklist. If the name of the service matches a blocked entry, the API Gateway Controller sets an appropriate validation status on the APIRule CR created for that service.

>**TIP:** For more information, read about the [Api CR statuses](#custom-resource-api-rule-status-codes).

The blocklist works as a security measure and prevents users from exposing vital internal services of Kubernetes, Istio, and API Server Proxy.
