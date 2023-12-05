---
title: Security in Kyma
---

To ensure a stable and secure work environment, the Kyma security component uses the following tools:

- Predefined [Kubernetes RBAC roles](https://kubernetes.io/docs/reference/access-authn-authz/rbac/) to manage the user access to the functionality provided by Kyma
- Istio Service Mesh with the global mTLS setup and ingress configuration to ensure secure service-to-service communication
- [ORY Oathkeeper](https://www.ory.sh/oathkeeper/docs/) used by API Gateway to authenticate, authorize, and mutate incoming HTTP requests.

This is a complete list of security-related guides in Kyma:

* [Authentication in Kyma](sec-01-authentication-in-kyma.md)
* [Authorization in Kyma](sec-02-authorization-in-kyma.md)
* [Access Kyma securely](sec-03-access-kyma.md)
