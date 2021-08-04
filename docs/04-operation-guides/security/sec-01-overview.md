---
title: Overview
---

To ensure a stable and secure work environment, the Kyma security component uses the following tools:

- Predefined [Kubernetes RBAC roles](https://kubernetes.io/docs/reference/access-authn-authz/rbac/) to manage the user access to the functionality provided by Kyma
- Istio Service Mesh with the global mTLS setup and ingress configuration to ensure secure service-to-service communication
- [ORY Oathkeeper](https://www.ory.sh/oathkeeper/docs/) and [ORY Hydra](https://www.ory.sh/hydra/docs/concepts/oauth2/) used by the API Gateway to authorize HTTP requests and provide the OAuth2 server functionality.
