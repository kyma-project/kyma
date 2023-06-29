---
title: Details
---

The main principle of Kyma Service Mesh is to inject Pods of every service with the Envoy sidecar proxy. Envoy intercepts the communication between the services and regulates it by applying and enforcing the rules you create.

By default, Istio in Kyma has [mutual TLS (mTLS)](https://istio.io/docs/concepts/security/#mutual-tls-authentication) disabled. See how to [enable sidecar proxy injection](../../04-operation-guides/operations/smsh-01-istio-enable-sidecar-injection.md). You can manage mTLS traffic in services or at a Namespace level by creating [DestinationRules](https://istio.io/docs/reference/config/networking/destination-rule/) and [Peer Authentications](https://istio.io/docs/tasks/security/authentication/authn-policy/). If you disable sidecar injection in a service or in a Namespace, you must manage their traffic configuration by creating appropriate DestinationRules and Peer Authentications.

> **NOTE:** The Istio Control Plane doesn't have mTLS enabled.

> **NOTE:** For security and performance we use the [distroless](https://istio.io/docs/ops/configuration/security/harden-docker-images/) version of Istio images. Those images are not Debian-based and are slimmed down to reduce any potential attack surface and increase startup time.

You can install Service Mesh as part of Kyma predefined [profiles](../../04-operation-guides/operations/02-install-kyma.md#choose-resource-consumption). For production purposes, use the **production profile** which has increased resource quotas for all Istio components. It also has HorizontalPodAutoscaler (HPA) enabled for all Istio components.
