---
title: Kiali Architecture
type: Architecture
---

The following diagram presents the overall Kiali architecture and the way the components interact with each other.

![Kiali architecture](assets/architecture.svg)

1. Use the Kyma Console or direct URL to access Kiali.
2. To ensure authentication, the Keycloak Gatekeeper checks if you have a valid token.
3. If not, the Keycloak Gatekeeper redirects you to the identity provider to log in.
4. After a successful log in, you can access the Kiali service which serves the website. The service exposes an endpoint and acts as an entry point for the Kiali deployment.
5. Kiali collects the information on the cluster health from the following sources:
  * API server which provides data on the cluster state.
  * Service Mesh by analyzing metrics Prometheus scrapes from the Istio Pod.
