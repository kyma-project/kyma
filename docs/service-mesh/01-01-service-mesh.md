---
title: Overview
---

Kyma Service Mesh is the component responsible for service-to-service communication, proxying, service discovery, traceability, and security. Kyma Service Mesh
is based on [Istio](https://istio.io/docs/concepts/what-is-istio/) open platform. The main principle of Kyma Service Mesh operation is the process of injecting Pods of every service with an Envoy - a sidecar proxy which intercepts the communication between the services and regulates it by applying and enforcing the rules you create. Kyma [Dex](https://github.com/coreos/dex), which is also a part of the Service Mesh, allows you to integrate any [OpenID Connect](https://openid.net/connect/)-compliant identity provider or a SAML2-based enterprise authentication server with your solution.

By default, the Kyma implementation of Istio has [mutual TLS (mTLS)](https://istio.io/docs/tasks/security/mutual-tls/) enabled and injects a sidecar container to every Pod. If you disable sidecar injection or mTLS traffic in a service or a Namespace, you must manage their traffic configuration by creating appropriate [Destination Rules](https://istio.io/docs/reference/config/networking/v1alpha3/destination-rule/) and [Authentication Policies](https://istio.io/docs/reference/config/istio.authentication.v1alpha1/). 

>**NOTE:** The Mesh Control Plane doesn't have mTLS enabled.

See this [Istio diagram](https://istio.io/docs/concepts/what-is-istio/arch.svg) to understand the relationship between the Istio components and services.
