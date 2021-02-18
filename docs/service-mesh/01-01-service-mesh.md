---
title: Overview
---

Kyma Service Mesh is the component responsible for service-to-service communication, proxying, service discovery, traceability, and security.
To deliver this functionality, Kyma Service Mesh uses [Istio](https://istio.io/docs/concepts/what-is-istio/) open platform.

The main principle of Kyma Service Mesh is to inject Pods of every service with the Envoy sidecar proxy. Envoy intercepts the communication between the services and regulates it by applying and enforcing the rules you create.
Kyma [Dex](https://github.com/dexidp/dex), which is also a part of the Service Mesh, allows you to integrate any [OpenID Connect](https://openid.net/connect/)-compliant identity provider or a SAML2-based enterprise authentication server with your solution.

By default, Istio in Kyma has [mutual TLS (mTLS)](https://istio.io/docs/concepts/security/#mutual-tls-authentication) enabled and injects a sidecar container to every Pod. You can manage mTLS traffic in services or at a Namespace level by creating [Destination Rules](https://istio.io/docs/reference/config/networking/destination-rule/) and [Peer Authentications](https://istio.io/docs/tasks/security/authentication/authn-policy/). If you disable sidecar injection in a service or in a Namespace, you must manage their traffic configuration by creating appropriate Destination Rules and Peer Authentications.

>**NOTE:** The Istio Control Plane doesn't have mTLS enabled.

>**NOTE:** For security and performance we use the [distroless](https://istio.io/docs/ops/configuration/security/harden-docker-images/) version of Istio images. Those images are not Debian-based and are slimmed down to reduce any potential attack surface and increase startup time.

You can install Service Mesh as part of Kyma predefined [profiles](/root/kyma/#installation-overview-profiles). For production purposes, use the **production profile** which has increased resource quotas for all Istio components. It also has Horizontal Pod Autoscaler (HPA) enabled for all Istio components.
