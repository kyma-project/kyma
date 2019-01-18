---
title: Overview
---

Kyma Service Mesh is the component responsible for service-to-service communication, proxying, service discovery, traceability, and security. Kyma Service Mesh
is based on [Istio](https://istio.io/docs/concepts/what-is-istio/overview.html) open platform. The main principle of Kyma Service Mesh operation is the process of injecting Pods of every service with an Envoy - a sidecar proxy which intercepts the communication between the services and regulates it by applying and enforcing the rules you create. Kyma [Dex](https://github.com/coreos/dex), which is also a part of the Service Mesh, allows you to integrate any [OpenID Connect](https://openid.net/connect/)-compliant identity provider or a SAML2-based enterprise authentication server with your solution.

See this [Istio diagram](https://istio.io/docs/concepts/what-is-istio/arch.svg) to understand the relationship between the Istio components and Services.
