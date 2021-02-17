---
title: Overview
---

To ensure a stable and secure way of extending your applications and creating functions or microservices, Kyma comes with a comprehensive set of tools that aim to mitigate any security issues, and, at the same time, enable a streamlined experience.

To ensure a safe work environment, the Kyma security component uses the following tools:

- Predefined [Kubernetes RBAC roles](https://kubernetes.io/docs/reference/access-authn-authz/rbac/) to manage the user access to the functionality provided by Kyma.
- Istio Service Mesh with the global mTLS setup and ingress configuration to ensure secure service-to-service communication.
- [Dex](https://github.com/dexidp/dex), a local identity provider with federation capabilities that acts as a proxy between the client application and external identity providers to allow seamless authorization. 
- [ORY Oathkeeper](https://www.ory.sh/oathkeeper/docs/) and [ORY Hydra](https://www.ory.sh/hydra/docs/concepts/oauth2/) used by the API Gateway to authorize HTTP requests, provide the OAuth2 server functionality and.

