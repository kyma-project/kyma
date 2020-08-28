---
title: Overview
---

To ensure stable and secure way of extending your applications and creating functions or microservices, Kyma comes with a comprehensive package of configured tools which aim to mitigate those issues, and enable a streamlined experience.

To ensure a safe work environment, the Kyma security component uses the following tools:

- Predefined [Kubernetes RBAC roles](https://kubernetes.io/docs/reference/access-authn-authz/rbac/) used to manage the user access to functionality provided by Kyma.
- Istio service mesh with global mTLS setup and ingress configuration used to ensure secure service-to-service communication.
- [Dex](https://github.com/dexidp/dex), a local identity provider with federation capabilities that acts as an additional layer between the client application and external identity providers. 
- [ORY Oathkeeper](https://www.ory.sh/oathkeeper/docs/) and [ORY Hydra](https://www.ory.sh/hydra/docs/concepts/oauth2/) that authorize HTTP requests and provide an OAuth2 server and API Gateway functionality.

