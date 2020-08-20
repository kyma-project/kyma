---
title: Overview
---

Breaking down a monolithic application into atomic services offers various benefits, including better agility, better scalability and better ability to reuse services. However, microservices also have particular security needs:
- To defend against man-in-the-middle attacks, they need traffic encryption.
- To provide flexible service access control, they need mutual TLS and fine-grained access policies.
- To determine who did what at what time, they need auditing tools.

![Security Overview](./assets/security-overview.svg)

Kyma Security provides a comprehensive package of configured security tools which aim to mitigate those issues, and enable a streamlined experience:
- Predefined kubernetes RBAC roles
- Istio service-mesh with global mTLS setup, Ingress configuration
- Dex as a local identity provider with federation capabilities
- Ory/Oathkeeper & Hydra providing a Oauth2 server and API Gateway functionality

