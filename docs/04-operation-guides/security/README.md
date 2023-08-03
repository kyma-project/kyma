---
title: Security in Kyma
---

To ensure a stable and secure work environment, the Kyma security component uses the following tools:

- Predefined [Kubernetes RBAC roles](https://kubernetes.io/docs/reference/access-authn-authz/rbac/) to manage the user access to the functionality provided by Kyma
- Istio Service Mesh with the global mTLS setup and ingress configuration to ensure secure service-to-service communication
- [ORY Oathkeeper](https://www.ory.sh/oathkeeper/docs/) and [ORY Hydra](https://www.ory.sh/hydra/docs/concepts/oauth2/) used by API Gateway to authorize HTTP requests and provide the OAuth2 server functionality.

This is a complete list of security-related guides in Kyma:

* [Authentication in Kyma](/04-operation-guides/security/sec-01-authentication-in-kyma.md)
* [Authorization in Kyma](/04-operation-guides/security/sec-02-authorization-in-kyma.md)
* [Access Kyma securely](/04-operation-guides/security/sec-03-access-kyma.md)
* [Ingress and Egress traffic](/04-operation-guides/security/sec-04-ingress-egress-traffic.md)
* [OAuth2 server customization and operations](/04-operation-guides/security/sec-05-customization-operation.md)
* [Access and Expose Grafana](/04-operation-guides/security/sec-06-access-expose-grafana.md)
* [Useful links](/04-operation-guides/security/sec-07-useful-links.md)
